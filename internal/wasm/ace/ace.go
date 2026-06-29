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

package ace

import (
	"fmt"
	"syscall/js"
)

type ACE struct {
	ace js.Value
}

func get() (*ACE, error) {
	ace := js.Global().Get("ace")
	if ace.IsUndefined() {
		return nil, fmt.Errorf("ace javascript not loaded")
	}
	ace.Get("config").Call("set", "basePath", "/resources/js/lib/ace")
	return &ACE{ace: ace}, nil
}

func (a *ACE) edit(id string) js.Value {
	return a.ace.Call("edit", id)
}

type Editor struct {
	ace    *ACE
	editor js.Value
}

func New(id string) (*Editor, error) {
	ace, err := get()
	if err != nil {
		return nil, err
	}
	return &Editor{
		ace:    ace,
		editor: ace.edit(id),
	}, nil
}

func (ed *Editor) SetValue(src string) {
	ed.editor.Call("setValue", src)
	ed.editor.Get("session").Get("selection").Call("clearSelection")
}

func (ed *Editor) GetValue() string {
	return ed.editor.Call("getValue").String()
}

func (ed *Editor) SetTheme(theme string) {
	ed.editor.Call("setTheme", theme)
}

func (ed *Editor) SetMode(mode string) {
	ed.editor.Get("session").Call("setMode", mode)
}
