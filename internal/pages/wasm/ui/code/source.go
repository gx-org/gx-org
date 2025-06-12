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
	"honnef.co/go/js/dom/v2"
)

type Source struct {
	code      *Code
	container *dom.HTMLDivElement
	input     *dom.HTMLDivElement
	control   *dom.HTMLDivElement

	lastSrc string
}

func newSource(code *Code, parent dom.Element) *Source {
	s := &Source{
		code:      code,
		container: code.gui.CreateDIV(parent, ui.Class("code_source_container")),
	}
	s.input = code.gui.CreateDIV(parent,
		ui.Class("code_source_textinput_container"),
		ui.Property("contenteditable", "true"),
		ui.Listener("input", s.onSourceChange),
		ui.Listener("keypress", s.onKeyPress),
		ui.Listener("keydown", s.onKeyDown),
	)
	s.input.AddEventListener("input", true, func(ev dom.Event) {
	})
	s.control = code.gui.CreateDIV(parent,
		ui.Class("code_source_controls_container"),
	)
	code.gui.CreateButton(s.control, "Run", s.onRun)
	return s
}

func (s *Source) onKeyDown(ev *dom.KeyboardEvent) {
	const tabKey = 9
	if ev.KeyCode() != tabKey {
		return
	}
	ev.PreventDefault()
	s.updateSource(insertTab)
}

func insertTab(src string, sel *ui.Selection) (string, bool) {
	srcLines := strings.Split(src, "\n")
	cursorLine := sel.Line()
	if cursorLine >= len(srcLines) {
		return src, false
	}
	currentLine := []rune(srcLines[cursorLine])
	cursorColumn := sel.Column()
	newLine := append([]rune{}, currentLine[:cursorColumn]...)
	newLine = append(newLine, []rune(tabSpaces)...)
	newLine = append(newLine, currentLine[cursorColumn:]...)
	srcLines[cursorLine] = string(newLine)
	sel.MoveColumnBy(tabSpaces)
	return strings.Join(srcLines, "\n"), true
}

func (s *Source) onKeyPress(ev *dom.KeyboardEvent) {
	if ev.ShiftKey() && ev.Key() == "Enter" {
		s.onRun(ev)
		ev.PreventDefault()
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

func (s *Source) set(src string) {
	s.lastSrc = src
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
}

func (s *Source) onRun(dom.Event) {
	go s.code.callAndWrite(s.code.runCode, s.lastSrc)
}

func (s *Source) updateSource(process func(src string, sel *ui.Selection) (string, bool)) {
	currentSrc := s.extractSource()
	sel := s.code.gui.CurrentSelection(s.input)
	currentSrc, cont := process(currentSrc, sel)
	if !cont {
		return
	}
	defer sel.SetAsCurrent()
	s.set(currentSrc)
	go s.code.callAndWrite(s.code.compileAndWrite, currentSrc)

}

func (s *Source) onSourceChange(dom.Event) {
	s.updateSource(func(src string, sel *ui.Selection) (string, bool) {
		return src, s.lastSrc != src
	})
}
