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

package main

import (
	"fmt"

	"github.com/gx-org/gx-org/internal/pages/wasm/lessons"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui/code"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui/text"
	"honnef.co/go/js/dom/v2"
)

func main() {
	gui := ui.New(dom.GetWindow())
	body, err := ui.FindElementByClass[dom.HTMLElement](gui, "root_container")
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return
	}

	textElement := text.New(gui, body)
	codeElement := code.New(gui, body)

	chapters, err := lessons.New()
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return
	}
	current := chapters[0].Content[0]
	textElement.SetContent(current)
	codeElement.SetContent(current)

	<-make(chan bool)
}
