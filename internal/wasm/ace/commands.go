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

type Commands struct {
	js.Value

	ed *Editor
}

func (ed *Editor) Commands() *Commands {
	return &Commands{
		Value: ed.Get("commands"),
		ed:    ed,
	}
}

type (
	KeyBinding struct {
		Win string
		Mac string
	}

	Command struct {
		Name    string
		BindKey KeyBinding
		Exec    func(*Editor)
	}
)

func (cmds *Commands) AddCommand(cmd Command) {
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		cmd.Exec(cmds.ed)
		return nil
	})
	cmds.Call("addCommand", map[string]any{
		"name": cmd.Name,
		"bindKey": map[string]any{
			"win": cmd.BindKey.Win,
			"mac": cmd.BindKey.Mac,
		},
		"exec": js.Func(cb),
	})
}
