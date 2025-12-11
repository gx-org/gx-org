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

package header

import (
	"github.com/gx-org/gx-org/internal/lessons"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Header struct {
	el dom.HTMLElement
}

func New(gui *ui.UI) *Header {
	els := gui.Dom().Document().GetElementsByTagName("header")
	hdr := &Header{}
	if len(els) == 0 {
		return hdr
	}
	hdr.el = els[0].(dom.HTMLElement)
	hdr.SetContent(nil)
	return hdr
}

func (hdr *Header) SetContent(les *lessons.Lesson) {
	pageTitle := "Walkthrough"
	if les != nil && les.Options.ChapterTitleAsPageTitle {
		pageTitle = les.Options.ChapterTitle
	}
	hdr.el.SetInnerHTML(`<a href="index.html"><img src="res/gxlogo.png" style="margin-left: 15px; margin-right: 15px; float: left; height:100%; object-fit: contain;" alt="GX Logo"></a>` + pageTitle)
}
