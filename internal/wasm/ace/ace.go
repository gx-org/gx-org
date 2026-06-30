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
	js.Value
}

func get() (*ACE, error) {
	ace := js.Global().Get("ace")
	if ace.IsUndefined() {
		return nil, fmt.Errorf("ace javascript not loaded")
	}
	ace.Get("config").Call("set", "basePath", "https://cdnjs.cloudflare.com/ajax/libs/ace/1.43.3/")
	return &ACE{Value: ace}, nil
}

func (a *ACE) edit(id string) js.Value {
	return a.Call("edit", id)
}
