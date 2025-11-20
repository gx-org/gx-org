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
	"fmt"
	"strconv"
	"strings"

	"github.com/gx-org/gx-org/internal/history"
	"github.com/gx-org/gx-org/internal/wasm/ui"
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

	keys      *ui.Keys
	source    *history.History[state]
	formatter *formatter
}

func newSource(code *Code, parent dom.Element) *Source {
	s := &Source{
		code:      code,
		container: code.gui.CreateDIV(parent, ui.Class("code_source_container")),
		source:    history.New(stateEq),
		formatter: newFormatter(),
	}
	s.input = code.gui.CreateDIV(parent,
		ui.Class("code_source_textinput_container"),
		ui.Property("contenteditable", "true"),
		ui.Listener("input", s.onSourceChange),
		ui.Listener("paste", s.onPaste),
		ui.KeyListener(s.onKeyPress),
	)
	s.control = code.gui.CreateDIV(parent,
		ui.Class("code_source_controls_container"),
	)
	code.gui.CreateButton(s.control, "Run", s.onRun)
	return s
}

func insertSource(src string, sel *ui.Selection, toInsert string) (string, *ui.Selection, bool) {
	cursorLine := sel.Line()
	var targetLines []string
	srcLines := strings.Split(src, "\n")
	for currentSrcLine, srcLine := range srcLines {
		if currentSrcLine < cursorLine {
			targetLines = append(targetLines, srcLine)
			continue
		}
		if currentSrcLine > cursorLine {
			targetLines = append(targetLines, srcLine)
			continue
		}
		srcLineRunes := []rune(srcLine)
		cursorColumn := min(sel.Column(), len(srcLineRunes))
		newLine := append([]rune{}, srcLineRunes[:cursorColumn]...)
		for insertedLine := range strings.Lines(toInsert) {
			newLine = append(newLine, []rune(strings.TrimSuffix(insertedLine, "\n"))...)
			if strings.HasSuffix(insertedLine, "\n") {
				targetLines = append(targetLines, string(newLine))
				newLine = []rune{}
				sel.MoveToNextLine()
			} else {
				sel.MoveColumnBy(insertedLine)
			}
		}
		newLine = append(newLine, srcLineRunes[cursorColumn:]...)
		targetLines = append(targetLines, string(newLine))
	}
	out := strings.Join(targetLines, "\n")
	return out, sel, true
}

func (s *Source) insertSource(inserted string) func(src string, sel *ui.Selection) (string, *ui.Selection, bool) {
	return func(src string, sel *ui.Selection) (string, *ui.Selection, bool) {
		return insertSource(src, sel, inserted)
	}
}

func (s *Source) onPaste(ev *dom.ClipboardEvent) {
	ev.PreventDefault()
	txt := ev.ClipboardData().GetData("text/plain")
	s.updateSource(s.insertSource(txt))
}

func (s *Source) onKeyPress(keys *ui.Keys, ev *dom.KeyboardEvent) {
	if keys.On("Shift") && keys.On("Enter") {
		// Run the code
		ev.PreventDefault()
		s.runCode()
		return
	}
	if keys.On("Enter") {
		ev.PreventDefault()
		s.updateSource(s.insertSource("\n"))
		return
	}
	if keys.On("Tab") {
		ev.PreventDefault()
		s.updateSource(s.insertSource(tabSpaces))
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
	src := ui.TextContent(s.input)
	if len(src) > 0 {
		src += "\n"
	}
	return src
}

const tabSize = 4

var tabSpaces = strings.Repeat(" ", tabSize)

func replaceNewLineWithBRTag(el dom.Node) {
	children := el.ChildNodes()
	if len(children) != 1 {
		return
	}
	data := children[0].Underlying().Get("data")
	if data.IsNull() || data.IsUndefined() {
		return
	}
	if data.String() != "\n" {
		return
	}
	dom.WrapHTMLElement(el.Underlying()).SetInnerHTML("<br>")
}

func setLineNumber(node dom.Node, lineNumber int) {
	dom.WrapElement(node.Underlying()).SetAttribute("line_num", strconv.Itoa(lineNumber))
}

func customizeChromaHTML() func(el dom.Node) bool {
	lineNumber := 0
	return func(el dom.Node) bool {
		if ui.NodeName(el) != "SPAN" {
			return true
		}
		for _, class := range ui.ClassOf(el) {
			switch class {
			case "chroma_w":
				replaceNewLineWithBRTag(el)
			case "chroma_line":
				setLineNumber(el, lineNumber)
				lineNumber++
			}
		}
		return true
	}
}

func (s *Source) set(src string, sel *ui.Selection) {
	s.source.Append(state{src: src, sel: sel})
	parent := s.input
	ui.ClearChildren(parent)
	parent.SetInnerHTML(s.formatter.format(src))
	ui.Walk(parent, customizeChromaHTML(), nil)
	if sel != nil {
		sel.SetAsCurrent()
	}
	ui.Go(func() {
		s.code.callAndWrite(s.code.compileAndWrite, src)
	})
}

func (s *Source) runCode() {
	s.code.callAndWrite(s.code.runCode, s.source.Current().src)
}

func (s *Source) onRun(dom.Event) {
	s.runCode()
}

func (s *Source) updateSource(process func(src string, sel *ui.Selection) (string, *ui.Selection, bool)) {
	currentSrc := s.extractSource()
	sel := s.code.gui.CurrentSelection(s.input)
	currentSrc, sel, cont := process(currentSrc, sel)
	if !cont {
		return
	}
	s.set(currentSrc, sel)
}

func (s *Source) onSourceChange(dom.Event) {
	s.updateSource(func(src string, sel *ui.Selection) (string, *ui.Selection, bool) {
		return src, sel, s.source.Current().src != src
	})
}
