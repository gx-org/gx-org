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
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type (
	Text struct {
		gui     *ui.UI
		page    Page
		lesson  *dom.HTMLDivElement
		content *dom.HTMLDivElement
		nav     *dom.HTMLDivElement
	}

	Page interface {
		DisplayLesson(*lessons.Lesson)
	}
)

func New(gui *ui.UI, parent dom.HTMLElement, page Page) *Text {
	text := &Text{
		gui:    gui,
		lesson: gui.CreateDIV(parent, ui.Class("lesson_container")),
		page:   page,
	}
	text.content = gui.CreateDIV(text.lesson, ui.Class("lesson_content"))
	text.nav = gui.CreateDIV(text.lesson, ui.Class("lesson_navigation"))
	return text
}

func (tt *Text) SetContent(les *lessons.Lesson) {
	ui.ClearChildren(tt.nav)
	tt.gui.CreateButton(tt.nav, "←",
		func(dom.Event) {
			tt.page.DisplayLesson(les.Prev)
		},
		ui.SetVisible(les.Prev != nil),
		ui.Class("navigation_button"),
	)
	tt.gui.CreateParagraph(tt.nav, fmt.Sprintf("Chapter %d, lesson %d/%d", les.Chapter.ID, les.ID, les.Chapter.NumLessons()))
	tt.gui.CreateButton(tt.nav, "→",
		func(dom.Event) {
			tt.page.DisplayLesson(les.Next)
		},
		ui.SetVisible(les.Next != nil),
		ui.Class("navigation_button"),
	)
	tt.content.SetInnerHTML(les.HTML)
}
