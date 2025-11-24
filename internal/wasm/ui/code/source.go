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

package code

import (
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"github.com/gx-org/gx-org/internal/wasm/ui/widgets/editor"
	"honnef.co/go/js/dom/v2"
)

type Source struct {
	code      *Code
	container *dom.HTMLDivElement
	control   *dom.HTMLDivElement
	editor    *editor.Editor
}

func newSource(code *Code, parent dom.Element) *Source {
	s := &Source{
		code:      code,
		container: code.gui.CreateDIV(parent, ui.Class("code_source_textinput_container")),
	}
	s.editor = editor.New(code.gui, s.container, s.updateCode, s.runCode)
	s.control = code.gui.CreateDIV(parent,
		ui.Class("code_source_controls_container"),
	)
	code.gui.CreateButton(s.control, "Run", func(dom.Event) {
		s.runCode(s.editor.Text())
	})
	return s
}

func (s *Source) updateCode(src string) {
	ui.Go(func() {
		s.code.updateCodeOutput(s.code.compile, src)
	})
}

func (s *Source) runCode(src string) {
	ui.Go(func() {
		s.code.updateCodeOutput(s.code.compileAndRun, src)
	})
}
