// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build wasm

package text

import (
	"fmt"

	"github.com/gx-org/gx-org/internal/lessons"
	"github.com/gx-org/gx-org/internal/wasm/codefmt"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type (
	Text struct {
		gui     *ui.UI
		page    Page
		lesson  *dom.HTMLDivElement
		content *dom.HTMLDivElement
		config  *dom.HTMLDivElement
		nav     *dom.HTMLDivElement

		configFmt *codefmt.Formatter
	}

	Page interface {
		DisplayLesson(*lessons.Lesson)
	}
)

func New(gui *ui.UI, parent dom.HTMLElement, page Page) *Text {
	text := &Text{
		gui:       gui,
		lesson:    gui.CreateDIV(parent, ui.Class("lesson_container")),
		page:      page,
		configFmt: codefmt.JSON(),
	}
	text.content = gui.CreateDIV(text.lesson, ui.Class("lesson_content"))
	text.config = gui.CreateDIV(
		gui.CreateDIV(text.lesson, ui.Class("lesson_config")),
		ui.Class("code_source_textinput"),
	)
	text.nav = gui.CreateDIV(text.lesson, ui.Class("lesson_navigation"))
	return text
}

func lessonFooter(les *lessons.Lesson) string {
	if les.Chapter.Name() == "intro" {
		return "Introduction"
	}
	return fmt.Sprintf("Chapter %d Lesson %d/%d", les.Chapter.ID, les.ID, les.Chapter.NumLessons())
}

func (tt *Text) setNavigation(les *lessons.Lesson) {
	ui.ClearChildren(tt.nav)
	tt.gui.CreateButton(tt.nav, "←",
		func(dom.Event) {
			tt.page.DisplayLesson(les.Prev)
		},
		ui.SetVisible(les.Prev != nil),
		ui.Class("navigation_button"),
	)
	tt.gui.CreateParagraph(tt.nav, lessonFooter(les))
	tt.gui.CreateButton(tt.nav, "→",
		func(dom.Event) {
			tt.page.DisplayLesson(les.Next)
		},
		ui.SetVisible(les.Next != nil),
		ui.Class("navigation_button"),
	)
}

func (tt *Text) setConfig(les *lessons.Lesson) {
	ui.ClearChildren(tt.config)
	if len(les.Config) == 0 {
		ui.SetVisibleProperty(tt.config, false)
		return
	} else {
		ui.SetVisibleProperty(tt.config, true)
	}
	config := tt.configFmt.Format(les.ConfigSrc)
	tt.config.SetInnerHTML("GX Config:<br>" + config)
}

func (tt *Text) SetContent(les *lessons.Lesson) {
	if les.Options.HideText {
		ui.SetVisibleProperty(tt.lesson, false)
		return
	} else {
		ui.SetVisibleProperty(tt.lesson, true)
	}
	tt.setConfig(les)
	tt.setNavigation(les)
	tt.content.SetInnerHTML(les.HTML)
}
