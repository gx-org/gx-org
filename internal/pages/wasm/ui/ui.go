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
	applyAll(el, opts)
	el.AddEventListener("click", true, func(ev dom.Event) {
		go f(ev)
	})
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

func findParentDIV(el js.Value) js.Value {
	current := el
	for !isDiv(current) {
		current = current.Get("parentElement")
	}
	return current
}

func lineNumFromElement(el js.Value) int {
	line := 0
	prev := findParentDIV(el).Get("previousElementSibling")
	for !prev.IsNull() {
		prev = prev.Get("previousElementSibling")
		line++
	}
	return line
}

func utf16Count(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func textLenFromPreviousElement(el js.Value) (utf16Pos, utf8Pos int) {
	if nodeName(el.Get("firstChild")) == "BR" {
		return
	}
	// Make sure that the parent is DIV
	// (moving up from text to font in a font tag)
	for !isDiv(el.Get("parentNode")) {
		el = el.Get("parentNode")
	}
	if el.Get("parentNode").Get("childNodes").Length() <= 1 {
		return
	}
	// Start counting text context in the previous element
	// (ignoring the current element)
	prev := el.Get("previousSibling")
	for !prev.IsNull() {
		text := TextContent(prev)
		utf16Pos += utf16Count(text)
		utf8Pos += utf8.RuneCountInString(text)
		prev = prev.Get("previousSibling")
	}
	return
}

func textLenFromElement(rang js.Value, ancestor js.Value) (utf16Pos, utf8Pos int) {
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
	ancestor := rang.Get("commonAncestorContainer")
	line := 0
	if len(el.InnerHTML()) > 1 { // Necessary condition to handle the edge case when there is only a single character.
		line = lineNumFromElement(ancestor)
	}
	utf16Prev, utf8Prev := textLenFromPreviousElement(ancestor)
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

func TextContent(el js.Value) string {
	var content strings.Builder
	for leaf := range iterLeaves(&dom.BasicNode{Value: el}) {
		data := leaf.Underlying().Get("data")
		if data.IsNull() || data.IsUndefined() {
			continue
		}
		content.WriteString(data.String())
	}
	return html.UnescapeString(content.String())
}

func iterLeaves(el dom.Node) func(yield func(dom.Node) bool) {
	return func(yield func(dom.Node) bool) {
		if !el.HasChildNodes() {
			yield(el)
			return
		}
		for _, child := range el.ChildNodes() {
			for leaf := range iterLeaves(child) {
				if !yield(leaf) {
					return
				}
			}
		}
	}
}

func findFirstLeaf(el dom.Node) dom.Node {
	for leaf := range iterLeaves(el) {
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
		textLen := utf16Count(TextContent(child.Underlying()))
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
