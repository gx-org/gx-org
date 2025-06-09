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

//go:build wasm

package text

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gx-org/gx-org/internal/pages/wasm/lessons"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Text struct {
	gui *ui.UI
	div *dom.HTMLDivElement
}

func New(gui *ui.UI, parent dom.HTMLElement) *Text {
	return &Text{
		gui: gui,
		div: gui.CreateDIV(parent, ui.Class("text_container")),
	}
}

func htmlFromMD(md string) string {
	p := parser.NewWithExtensions(parser.CommonExtensions)
	doc := p.Parse([]byte(md))
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func (tt *Text) SetContent(les *lessons.Lesson) {
	tt.div.SetInnerHTML(htmlFromMD(les.Text))
}
