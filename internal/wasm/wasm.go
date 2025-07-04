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
	"net/url"
	"strconv"

	"github.com/gx-org/gx-org/internal/lessons"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"github.com/gx-org/gx-org/internal/wasm/ui/code"
	"github.com/gx-org/gx-org/internal/wasm/ui/text"
	"honnef.co/go/js/dom/v2"
)

type root struct {
	gui  *ui.UI
	text *text.Text
	code *code.Code
}

func (r *root) DisplayLesson(les *lessons.Lesson) {
	r.text.SetContent(les)
	r.code.SetContent(les)
	r.gui.UpdateURL(fmt.Sprintf("index.html?chapter=%s&lesson=%d", les.Chapter.Name(), les.ID))
}

func idsFromURL(loc *url.URL) (string, int) {
	chapS := loc.Query().Get("chapter")
	if chapS == "" {
		return "", 0
	}
	lesS := loc.Query().Get("lesson")
	if lesS == "" {
		return chapS, 0
	}
	lesID, _ := strconv.Atoi(lesS)
	return chapS, lesID
}

func main() {
	gui := ui.New(dom.GetWindow())
	body, err := ui.FindElementByClass[dom.HTMLElement](gui, "root_container")
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return
	}

	root := &root{gui: gui}
	root.text = text.New(gui, body, root)
	root.code = code.New(gui, body)

	chapters, err := lessons.New()
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return
	}
	loc, err := gui.URL()
	var chapName string
	var lessonID int
	if err != nil {
		fmt.Println("URL ERROR", err.Error())
	} else {
		chapName, lessonID = idsFromURL(loc)
	}
	root.DisplayLesson(lessons.FindLesson(chapters, chapName, lessonID))

	<-make(chan bool)
}
