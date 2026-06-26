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

type Text struct {
	gui     *ui.UI
	lesson  *dom.HTMLDivElement
	content *dom.HTMLDivElement
	config  *dom.HTMLDivElement
	nav     *dom.HTMLDivElement

	configFmt *codefmt.Formatter
	actions   navActions
}

func New(gui *ui.UI, parent dom.HTMLElement, page Page) *Text {
	text := &Text{
		gui: gui,
		lesson: gui.CreateDIV(parent,
			ui.Class("lesson_container"),
			ui.TabIndex(0),
			ui.Focus(),
		),
		configFmt: codefmt.JSON(),
		actions:   navActions{page: page},
	}
	text.content = gui.CreateDIV(text.lesson, ui.Class("lesson_content"))
	text.config = gui.CreateDIV(
		gui.CreateDIV(text.lesson, ui.Class("lesson_config")),
		ui.Class("code_source_textinput"),
	)
	text.nav = gui.CreateDIV(text.lesson, ui.Class("lesson_navigation"))
	text.lesson.AddEventListener("keydown", true, text.onKeyDown)
	return text
}

func lessonFooter(les *lessons.Lesson) string {
	return fmt.Sprintf("%s. Lesson %d/%d", les.Chapter.Title, les.ID, les.Chapter.NumLessons())
}

func (tt *Text) onKeyDown(ev dom.Event) {
	keyEvent, isKeyEvent := ev.(*dom.KeyboardEvent)
	if !isKeyEvent {
		return
	}
	switch keyEvent.Key() {
	case "ArrowLeft":
		tt.actions.displayPrevious()
	case "ArrowRight":
		tt.actions.displayNext()
	}
}

func (tt *Text) setNavigation() {
	ui.ClearChildren(tt.nav)
	tt.gui.CreateButton(tt.nav, "←",
		func(dom.Event) {
			tt.actions.displayPrevious()
		},
		ui.SetVisible(tt.actions.hasPrev()),
		ui.Class("navigation_button"),
	)
	tt.gui.CreateParagraph(tt.nav, lessonFooter(tt.actions.current))
	tt.gui.CreateButton(tt.nav, "→",
		func(dom.Event) {
			tt.actions.displayNext()
		},
		ui.SetVisible(tt.actions.hasNext()),
		ui.Class("navigation_button"),
	)
}

func (tt *Text) setConfig() {
	les := tt.actions.current
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
	tt.actions.current = les
	if les.Options.HideText {
		ui.SetVisibleProperty(tt.lesson, false)
		return
	} else {
		ui.SetVisibleProperty(tt.lesson, true)
	}
	tt.setConfig()
	tt.setNavigation()
	tt.content.SetInnerHTML(les.HTML)
}
