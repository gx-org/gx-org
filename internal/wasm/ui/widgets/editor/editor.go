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
	el                  *dom.HTMLDivElement
	ace                 *ace.Editor
	updateCode, runCode func(string)
}

func New(gui *ui.UI, parent dom.Element, updateCode, runCode func(src string)) *Editor {
	const id = "editor"
	ed := &Editor{
		updateCode: updateCode,
		runCode:    runCode,
	}
	ed.el = gui.CreateDIV(
		parent,
		ui.ID(id),
	)
	var err error
	ed.ace, err = ace.New(id)
	if err != nil {
		fmt.Println("ERROR", err.Error())
	}
	ed.ace.SetTheme("ace/theme/xcode")
	ed.ace.Session().SetMode("ace/mode/golang")
	ed.ace.Session().OnChange(ed.onChange)
	ed.ace.Commands().AddCommand(ace.Command{
		Name: "Run",
		BindKey: ace.KeyBinding{
			Mac: "Shift-Enter",
			Win: "Shift-Enter",
		},
		Exec: ed.onRunCode,
	})
	return ed
}

func (ed *Editor) onRunCode(*ace.Editor) {
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
