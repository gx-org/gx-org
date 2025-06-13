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
	"html"
	"strings"

	"github.com/gx-org/gx-org/internal/pages/wasm/ui"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui/history"
	"honnef.co/go/js/dom/v2"
)

type state struct {
	src string
	sel *ui.Selection
}

func (s state) String() string {
	line, column := -1, -1
	if s.sel != nil {
		line, column = s.sel.Line(), s.sel.Column()
	}
	return fmt.Sprintf("%d:%d:%s", line, column, s.src)
}

func stateEq(a, b state) bool {
	return a.src == b.src
}

type Source struct {
	code      *Code
	container *dom.HTMLDivElement
	input     *dom.HTMLDivElement
	control   *dom.HTMLDivElement

	keys   *ui.Keys
	source *history.History[state]
}

func newSource(code *Code, parent dom.Element) *Source {
	s := &Source{
		code:      code,
		container: code.gui.CreateDIV(parent, ui.Class("code_source_container")),
		source:    history.New[state](stateEq),
	}
	s.input = code.gui.CreateDIV(parent,
		ui.Class("code_source_textinput_container"),
		ui.Property("contenteditable", "true"),
		ui.Listener("input", s.onSourceChange),
		ui.KeyListener(s.onKeyPress),
	)
	s.control = code.gui.CreateDIV(parent,
		ui.Class("code_source_controls_container"),
	)
	code.gui.CreateButton(s.control, "Run", s.onRun)
	return s
}

func insertTab(src string, sel *ui.Selection) (string, *ui.Selection, bool) {
	srcLines := strings.Split(src, "\n")
	cursorLine := sel.Line()
	if cursorLine >= len(srcLines) {
		return src, nil, false
	}
	currentLine := []rune(srcLines[cursorLine])
	cursorColumn := sel.Column()
	newLine := append([]rune{}, currentLine[:cursorColumn]...)
	newLine = append(newLine, []rune(tabSpaces)...)
	newLine = append(newLine, currentLine[cursorColumn:]...)
	srcLines[cursorLine] = string(newLine)
	sel.MoveColumnBy(tabSpaces)
	return strings.Join(srcLines, "\n"), sel, true
}

func (s *Source) onKeyPress(keys *ui.Keys, ev *dom.KeyboardEvent) {
	if keys.On("Shift", "Enter") {
		s.onRun(ev)
		ev.PreventDefault()
		return
	}
	if keys.On("Tab") {
		ev.PreventDefault()
		s.updateSource(insertTab)
		return
	}
	if (keys.On("Meta") || keys.On("Control")) && keys.On("z") {
		s.updateSource(func(string, *ui.Selection) (string, *ui.Selection, bool) {
			if keys.On("Shift") {
				s.source.Redo()
			} else {
				s.source.Undo()
			}
			current := s.source.Current()
			return current.src, current.sel, true
		})
		ev.PreventDefault()
		return
	}
}

func (s *Source) extractSource() string {
	var srcs []string
	for _, child := range s.input.ChildNodes() {
		srcs = append(srcs, ui.TextContent(child.Underlying()))
	}
	src := strings.Join(srcs, "\n")
	src = strings.ReplaceAll(src, "\u00a0", " ")
	return src
}

var keywordToColor = []struct {
	color string
	words []string
}{
	{
		color: "var(--language-keyword)",
		words: []string{
			"var", "const", "return", "struct", "func", "package", "import",
		},
	},
	{
		color: "var(--type-keyword)",
		words: []string{
			"bool", "string",
			"int32", "int64",
			"bfloat64", "float32", "float64",
		},
	},
}

const tabSize = 4

var tabSpaces = strings.Repeat(" ", tabSize)

func format(s string) string {
	s = strings.ReplaceAll(s, "\t", tabSpaces)
	s = strings.ReplaceAll(s, " ", "\u00a0")
	s = html.EscapeString(s)
	for _, color := range keywordToColor {
		fontTag := fmt.Sprintf(`<span style="color:%s;">%%s</span>`, color.color)
		for _, word := range color.words {
			s = strings.ReplaceAll(s, word, fmt.Sprintf(fontTag, word))
		}
	}
	return s
}

func (s *Source) set(src string, sel *ui.Selection) {
	s.source.Append(state{src: src, sel: sel})
	parent := s.input
	ui.ClearChildren(parent)
	for _, line := range strings.Split(src, "\n") {
		if line == "" {
			line = "<br>"
		} else {
			line = format(line)
		}
		s.code.gui.CreateDIV(parent,
			ui.InnerHTML(line),
		)
	}
	if sel != nil {
		sel.SetAsCurrent()
	}
}

func (s *Source) onRun(dom.Event) {
	go s.code.callAndWrite(s.code.runCode, s.source.Current().src)
}

func (s *Source) updateSource(process func(src string, sel *ui.Selection) (string, *ui.Selection, bool)) {
	currentSrc := s.extractSource()
	sel := s.code.gui.CurrentSelection(s.input)
	currentSrc, sel, cont := process(currentSrc, sel)
	if !cont {
		return
	}
	s.set(currentSrc, sel)
	go s.code.callAndWrite(s.code.compileAndWrite, currentSrc)

}

func (s *Source) onSourceChange(dom.Event) {
	s.updateSource(func(src string, sel *ui.Selection) (string, *ui.Selection, bool) {
		return src, sel, s.source.Current().src != src
	})
}
