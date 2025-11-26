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

func (ui *UI) Dom() dom.Window {
	return ui.win
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

func ClearChildren(node dom.Node) {
	for _, child := range node.ChildNodes() {
		node.RemoveChild(child)
	}
}

func NodeName(el dom.Node) string {
	return strings.ToUpper(el.NodeName())
}

func ClassOf(node dom.Node) []string {
	if node.Underlying().Get("classList").IsUndefined() {
		return nil
	}
	classes := node.(dom.Element).Class()
	ss := make([]string, classes.Length())
	for i := range ss {
		ss[i] = classes.Item(i)
	}
	return ss
}

func Walk(el dom.Node, onAll func(dom.Node) bool, onLeaf func(dom.Node, string) bool) bool {
	if onAll != nil && !onAll(el) {
		return false
	}
	if onLeaf != nil && !el.HasChildNodes() {
		data := el.Underlying().Get("data")
		text := ""
		if !data.IsNull() && !data.IsUndefined() {
			text = data.String()
		}
		if !onLeaf(el, text) {
			return false
		}
	}
	for _, child := range el.ChildNodes() {
		if !Walk(child, onAll, onLeaf) {
			return false
		}
	}
	return true
}

func IterLeaves(el dom.Node, cancel func(dom.Node) bool) func(yield func(dom.Node, string) bool) {
	return func(yield func(dom.Node, string) bool) {
		Walk(el, cancel, yield)
	}
}
