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
	var pathPrefix string
	if les.Chapter.PathPrefix != "" {
		pathPrefix = fmt.Sprintf("&pathPrefix=%s", les.Chapter.PathPrefix)
	}
	r.gui.UpdateURL(fmt.Sprintf("index.html?chapter=%d&lesson=%d%s", les.Chapter.ID, les.ID, pathPrefix))
}

func idsFromURL(loc *url.URL) (string, int, int) {
	pathPrefix := loc.Query().Get("prefix")
	chapID, _ := strconv.Atoi(loc.Query().Get("chapter"))
	lesID, _ := strconv.Atoi(loc.Query().Get("lesson"))
	return pathPrefix, chapID, lesID
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
	var pathPrefix string
	var chapID, lessonID int
	if err != nil {
		fmt.Println("URL ERROR", err.Error())
	} else {
		pathPrefix, chapID, lessonID = idsFromURL(loc)
	}
	root.DisplayLesson(lessons.FindLesson(chapters, pathPrefix, chapID, lessonID))

	<-make(chan bool)
}
