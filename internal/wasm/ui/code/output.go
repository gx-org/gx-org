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

package code

import (
	"fmt"

	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Output struct {
	code *Code
	div  *dom.HTMLDivElement

	crash string
}

func newOutput(code *Code, parent dom.Element) *Output {
	out := &Output{
		code: code,
		div:  code.gui.CreateDIV(parent, ui.Class("code_output_container")),
	}
	ui.SetCrashOutput(func(crash string) {
		out.crash = crash
		out.set("")
	})
	return out
}

func (out *Output) set(src string) {
	crash := ""
	if out.crash != "" {
		crash = "\n" + out.crash
	}
	out.div.SetInnerHTML(fmt.Sprintf("<pre>%s%s<pre>", src, crash))
}
