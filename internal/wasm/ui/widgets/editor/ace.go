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

type ACEEditor struct {
	el                  *dom.HTMLDivElement
	ace                 *ace.Editor
	updateCode, runCode func(string)
}

func New(gui *ui.UI, parent dom.Element, updateCode, runCode func(src string)) *ACEEditor {
	const id = "editor"
	editor := &ACEEditor{
		updateCode: updateCode,
		runCode:    runCode,
		el: gui.CreateDIV(
			parent,
			ui.ID(id),
		),
	}
	var err error
	editor.ace, err = ace.New(id)
	if err != nil {
		fmt.Println("ERROR", err.Error())
	}
	editor.ace.SetTheme("ace/theme/xcode")
	editor.ace.SetMode("ace/mode/golang")
	return editor
}

func (ed *ACEEditor) Set(src string) {
	ed.ace.SetValue(src)
}

func (ed *ACEEditor) Text() string {
	return ed.ace.GetValue()
}
