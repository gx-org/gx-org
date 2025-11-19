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

package ui

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"syscall/js"
	"unicode/utf16"
	"unicode/utf8"

	"honnef.co/go/js/dom/v2"
)

type UI struct {
	win dom.Window
}

func New(win dom.Window) *UI {
	return &UI{win}
}

func (ui *UI) UpdateURL(newURL string) {
	js.Global().Get("window").Get("history").Call("pushState", nil, "GX", newURL)
}

func (ui *UI) URL() (*url.URL, error) {
	return url.Parse(ui.win.Location().Href())
}

func (ui *UI) CreateDIV(parent dom.Element, opts ...ElementOption) *dom.HTMLDivElement {
	el := ui.win.Document().CreateElement("div")
	parent.AppendChild(el)
	applyAll(el, opts)
	return el.(*dom.HTMLDivElement)
}

func (ui *UI) CreateBR(parent dom.Element, opts ...ElementOption) *dom.HTMLBRElement {
	el := ui.win.Document().CreateElement("br")
	parent.AppendChild(el)
	applyAll(el, opts)
	return el.(*dom.HTMLBRElement)
}

type EventFunc func(ev dom.Event)

func (ui *UI) CreateButton(parent dom.Element, text string, f EventFunc, opts ...ElementOption) *dom.HTMLButtonElement {
	el := ui.win.Document().CreateElement("button")
	el.SetTextContent(text)
	parent.AppendChild(el)
	applyAll(el, []ElementOption{
		Listener("click", func(ev dom.Event) {
			Go(func() {
				f(ev)
			})
		}),
	})
	applyAll(el, opts)
	return el.(*dom.HTMLButtonElement)
}

func (ui *UI) CreateParagraph(parent dom.Element, text string, opts ...ElementOption) *dom.HTMLParagraphElement {
	el := ui.win.Document().CreateElement("p")
	parent.AppendChild(el)
	el.SetInnerHTML(html.EscapeString(text))
	return el.(*dom.HTMLParagraphElement)
}

func FindElementByClass[T dom.Element](ui *UI, class string) (zero T, err error) {
	els := ui.win.Document().GetElementsByClassName(class)
	if len(els) == 0 {
		return zero, fmt.Errorf("not element of class %s found", class)
	}
	if len(els) > 1 {
		return zero, fmt.Errorf("too many elements of class %s found", class)
	}
	el := els[0]
	elT, ok := el.(T)
	if !ok {
		return zero, fmt.Errorf("node %s:%T cannot be converted %T", el, el, zero)
	}
	return elT, nil
}

type Selection struct {
	ui          *UI
	el          dom.HTMLElement
	line        int
	utf16Column int
	utf8Column  int
	rang        js.Value
}

func selection() js.Value {
	return js.Global().Call("getSelection")
}

func nodeName(el js.Value) string {
	if el.IsNull() {
		return ""
	}
	return strings.ToUpper(el.Get("nodeName").String())
}

func isDiv(el js.Value) bool {
	return nodeName(el) == "DIV"
}

func findParentLineAndCode(el dom.Element) (line, code dom.Element) {
	current := el
	for current != nil && NodeName(current) != "CODE" {
		for _, class := range ClassOf(current) {
			if strings.HasSuffix(class, "line") {
				line = current
			}
		}
		current = current.ParentElement()
	}
	code = current
	return
}

func lineNumFromElement(el dom.Element) (dom.Element, int) {
	lineEl, codeEl := findParentLineAndCode(el)
	if codeEl == nil || lineEl == nil {
		return nil, 0
	}
	line := 0
	for i, child := range codeEl.ChildNodes() {
		if child.IsEqualNode(lineEl) {
			line = i + 1
		}
	}
	return lineEl, line
}

func utf16Count(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func textLenFromPreviousElement(ancestor dom.Element, line dom.Element) (utf16Pos, utf8Pos int) {
	text := TextContentUntil(line, func(el dom.Node) bool {
		return !el.IsEqualNode(ancestor)
	})
	return utf16Count(text), utf8.RuneCountInString(text)
}

func textLenFromElement(rang js.Value, ancestor dom.Element) (utf16Pos, utf8Pos int) {
	utf16Pos = rang.Get("startOffset").Int()
	utf8Str := TextContent(ancestor)
	utf16Str := utf16.Encode([]rune(utf8Str))
	if utf16Pos > len(utf16Str) {
		return 0, 0
	}
	subUTF16 := utf16Str[:utf16Pos]
	utf8Pos = len(utf16.Decode(subUTF16))
	return
}

func (ui *UI) CurrentSelection(el dom.HTMLElement) *Selection {
	if numRange := selection().Get("rangeCount").Int(); numRange == 0 {
		return nil
	}
	rang := selection().Call("getRangeAt", 0)
	ancestor := dom.WrapElement(rang.Get("commonAncestorContainer"))
	lineEl, line := lineNumFromElement(ancestor)
	utf16Prev, utf8Prev := textLenFromPreviousElement(ancestor, lineEl)
	utf16Column, utf8Column := textLenFromElement(rang, ancestor)
	return &Selection{
		ui:          ui,
		el:          el,
		rang:        rang,
		utf16Column: utf16Prev + utf16Column,
		utf8Column:  utf8Prev + utf8Column,
		line:        line,
	}
}

func noFilter(dom.Node) bool {
	return true
}

func TextContent(el dom.Node) string {
	return TextContentUntil(el, noFilter)
}

func TextContentUntil(el dom.Node, filter func(dom.Node) bool) string {
	var content strings.Builder
	for leaf := range iterLeaves(el, filter) {
		data := leaf.Underlying().Get("data")
		if data.IsNull() || data.IsUndefined() {
			continue
		}
		content.WriteString(data.String())
	}
	return html.UnescapeString(content.String())
}

func iterLeaves(el dom.Node, filter func(dom.Node) bool) func(yield func(dom.Node) bool) {
	return func(yield func(dom.Node) bool) {
		if !filter(el) {
			return
		}
		if !el.HasChildNodes() {
			yield(el)
			return
		}
		for _, child := range el.ChildNodes() {
			if !filter(child) {
				return
			}
			for leaf := range iterLeaves(child, filter) {
				if !yield(leaf) {
					return
				}
			}
		}
	}
}

func findFirstLeaf(el dom.Node) dom.Node {
	for leaf := range iterLeaves(el, noFilter) {
		return leaf
	}
	return nil
}

func (sel *Selection) SetAsCurrent() {
	if sel == nil {
		return
	}
	children := sel.el.ChildNodes()
	if sel.line >= len(children) {
		return
	}
	lineDiv := children[sel.line]
	column := sel.utf16Column
	for _, child := range lineDiv.ChildNodes() {
		textLen := utf16Count(TextContent(child))
		if column <= textLen {
			selection().Call("collapse", findFirstLeaf(child).Underlying(), column)
			return
		}
		column -= textLen
	}
}

func (sel *Selection) Line() int {
	return sel.line
}

func (sel *Selection) Column() int {
	return sel.utf8Column
}

func (sel *Selection) MoveColumnBy(s string) {
	sel.utf16Column += utf16Count(s)
	sel.utf8Column += utf8.RuneCountInString(s)
}

func (sel *Selection) MoveToNextLine() {
	sel.utf16Column = 0
	sel.utf8Column = 0
	sel.line++
}

func (sel *Selection) String() string {
	if sel == nil {
		return "nil"
	}
	return fmt.Sprintf("line: %d col: %d", sel.line, sel.utf16Column)
}

func ClearChildren(node dom.Node) {
	for _, child := range node.ChildNodes() {
		node.RemoveChild(child)
	}
}

func NodeName(el dom.Node) string {
	return strings.ToUpper(el.NodeName())
}

func ClassOf(el dom.Element) []string {
	if el.Underlying().Get("classList").IsUndefined() {
		return nil
	}
	classes := el.Class()
	ss := make([]string, classes.Length())
	for i := range ss {
		ss[i] = classes.Item(i)
	}
	return ss
}
