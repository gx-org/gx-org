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

//go:build js && wasm

package editor

import (
	"fmt"

	"github.com/gx-org/gx-org/internal/wasm/ace"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Editor struct {
	ace                 *ace.Editor
	updateCode, runCode func(string)
}

func New(gui *ui.UI, parent dom.Element, updateCode, runCode func(src string)) *Editor {
	const id = "editor"
	ed := &Editor{
		updateCode: updateCode,
		runCode:    runCode,
	}
	gui.CreateDIV(parent, ui.ID(id))
	var err error
	ed.ace, err = ace.New(id)
	if err != nil {
		fmt.Println("ERROR", err.Error())
	}
	ed.ace.SetTheme("ace/theme/xcode")
	ed.ace.Session().SetMode("ace/mode/golang")
	ed.ace.Session().OnChange(ed.onChange)
	// Run the code on Shift+Enter wherever the focus is: in the editor or
	// anywhere else in the page (e.g. main_root or lesson_container).
	gui.Dom().Document().AddEventListener("keydown", false, ed.onKeyDown)
	return ed
}

// onKeyDown runs the code when Shift+Enter is pressed anywhere in the document,
// unless the key has already been handled.
func (ed *Editor) onKeyDown(ev dom.Event) {
	keyEvent, isKeyEvent := ev.(*dom.KeyboardEvent)
	if !isKeyEvent || keyEvent.DefaultPrevented() {
		return
	}
	if keyEvent.Key() != "Enter" || !keyEvent.ShiftKey() ||
		keyEvent.CtrlKey() || keyEvent.AltKey() || keyEvent.MetaKey() {
		return
	}
	// Prevent the default action, for example inserting a new line in the
	// editor or clicking on a focused button.
	keyEvent.PreventDefault()
	ed.runCode(ed.Text())
}

func (ed *Editor) onChange() {
	ed.updateCode(ed.Text())
}

func (ed *Editor) Set(src string) {
	ed.ace.SetValue(src)
}

func (ed *Editor) Text() string {
	return ed.ace.GetValue()
}
