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

import "syscall/js"

type Editor struct {
	js.Value
	ace *ACE
}

func New(id string) (*Editor, error) {
	ace, err := get()
	if err != nil {
		return nil, err
	}
	return &Editor{
		ace:   ace,
		Value: ace.edit(id),
	}, nil
}

func (ed *Editor) SetValue(src string) {
	ed.Call("setValue", src)
	ed.Session().Get("selection").Call("clearSelection")
}

func (ed *Editor) GetValue() string {
	return ed.Call("getValue").String()
}

func (ed *Editor) SetTheme(theme string) {
	ed.Call("setTheme", theme)
}
