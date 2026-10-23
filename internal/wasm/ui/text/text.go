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

type Text struct {
	gui     *ui.UI
	lesson  *dom.HTMLDivElement
	content *dom.HTMLDivElement
	nav     *dom.HTMLDivElement

	actions navActions
}

func New(gui *ui.UI, parent dom.HTMLElement, page Page) *Text {
	text := &Text{
		gui: gui,
		lesson: gui.CreateDIV(parent,
			ui.Class("lesson_container"),
			ui.TabIndex(0),
			ui.Focus(),
		),
		actions: navActions{page: page},
	}
	text.content = gui.CreateDIV(text.lesson, ui.Class("lesson_content"))
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

// SVG paths of the navigation arrows in a 24x24 view box: filled triangles
// with their base on the side of the text. The fill color is defined in
// res/style.css.
const (
	arrowLeft  = "M19 5v14L8 12z"
	arrowRight = "M5 5v14l11-7z"
)

func (tt *Text) createNavButton(label, arrow string, enabled bool, onClick func()) {
	button := tt.gui.CreateButton(tt.nav, "",
		func(dom.Event) {
			onClick()
		},
		ui.Class("navigation_button"),
	)
	button.SetAttribute("aria-label", label)
	button.SetTitle(label)
	// A disabled button is hidden by the CSS but keeps its space
	// so that the text stays centered.
	button.SetDisabled(!enabled)
	tt.gui.CreateSVGIcon(button, "0 0 24 24", arrow)
}

func (tt *Text) setNavigation() {
	ui.ClearChildren(tt.nav)
	tt.createNavButton("Previous lesson", arrowLeft, tt.actions.hasPrev(), tt.actions.displayPrevious)
	tt.gui.CreateParagraph(tt.nav, lessonFooter(tt.actions.current))
	tt.createNavButton("Next lesson", arrowRight, tt.actions.hasNext(), tt.actions.displayNext)
}

// Nav returns the navigation div so the caller can reposition it in the DOM.
func (tt *Text) Nav() *dom.HTMLDivElement {
	return tt.nav
}

func (tt *Text) SetContent(les *lessons.Lesson) {
	tt.actions.current = les
	if les.Options.HideText {
		ui.SetVisibleProperty(tt.lesson, false)
		return
	} else {
		ui.SetVisibleProperty(tt.lesson, true)
	}
	tt.setNavigation()
	tt.content.SetInnerHTML(les.HTML)
}
