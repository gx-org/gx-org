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

package lessons

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gx-org/gx-org/internal/mdtext"
	"github.com/gx-org/gx-org/lessons"
)

type (
	Options struct {
		HideText                bool   `json:"hide_text"`
		ChapterTitle            string `json:"chapter_title"`
		ChapterTitleAsPageTitle bool   `json:"chapter_title_as_page_title"`
	}

	Chapter struct {
		name string

		PathPrefix string
		ID         int
		Title      string
		Content    []*Lesson
		Prev       *Chapter
		Next       *Chapter
	}

	Lesson struct {
		Chapter  *Chapter
		FileName string
		ID       int

		HTML      string
		Code      string
		Options   Options
		Config    map[string]any
		ConfigSrc string

		Prev *Lesson
		Next *Lesson
	}
)

var chapters = map[string][]string{
	"": {
		"intro",
		"types",
		"func",
		"struct",
	},
	"devs": {
		"fill",
		"demo",
	},
	"demos": {
		"reverse",
	},
}

func New() (map[string][]*Chapter, error) {
	m := make(map[string][]*Chapter)
	for pathPrefix, chapterNames := range chapters {
		var err error
		m[pathPrefix], err = readChapters(pathPrefix, chapterNames)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func readChapters(pathPrefix string, chapterNames []string) ([]*Chapter, error) {
	chapters := make([]*Chapter, len(chapterNames))
	var last *Lesson
	for i, name := range chapterNames {
		chap := &Chapter{name: name, ID: i, PathPrefix: pathPrefix}
		chapters[i] = chap
		var err error
		last, err = chap.readLessons(last)
		if err != nil {
			return nil, err
		}
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no content found")
	}
	return chapters, nil
}

func (chap *Chapter) readLessons(prevL *Lesson) (*Lesson, error) {
	for {
		lesson, err := chap.readLesson()
		if err != nil {
			return nil, err
		}
		if lesson == nil {
			return prevL, nil
		}
		if prevL != nil {
			lesson.Prev = prevL
			prevL.Next = lesson
		}
		prevL = lesson
	}
}

const defaultCode = `package main
`

func (chap *Chapter) readLesson() (*Lesson, error) {
	lessonID := len(chap.Content) + 1
	fileName := fmt.Sprintf("%s_%d.mdl", chap.Name(), lessonID)
	path := fileName
	if chap.PathPrefix != "" {
		path = filepath.Join(chap.PathPrefix, path)
	}
	data, err := lessons.Lessons.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) && lessonID > 1 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", path, err)
	}
	mdt := mdtext.Parse(data)
	lesson := &Lesson{
		Chapter:  chap,
		FileName: fileName,
		ID:       lessonID,
		Config:   make(map[string]any),
	}
	lesson.HTML = mdt.HTML
	lesson.Code = mdt.Code["main"]
	if lesson.Code == "" {
		lesson.Code = defaultCode
	}
	chap.Content = append(chap.Content, lesson)
	if err := lesson.parseOptions(mdt); err != nil {
		return nil, err
	}
	if chap.Title == "" {
		chap.Title = fmt.Sprintf("Chapter %d", chap.ID)
	}
	if lesson.Options.ChapterTitle != "" {
		chap.Title = lesson.Options.ChapterTitle
	}
	if err := lesson.parseConfig(mdt); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (chap *Chapter) Name() string {
	return chap.name
}

func (chap *Chapter) NumLessons() int {
	return len(chap.Content)
}

func get[T any](i int, ts []T) T {
	if i < 0 {
		return ts[0]
	}
	if i >= len(ts) {
		return ts[len(ts)-1]
	}
	return ts[i]
}

func (l *Lesson) parseOptions(mdt *mdtext.MDText) error {
	options := mdt.Code["options"]
	if options == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(options), &(l.Options)); err != nil {
		return fmt.Errorf("%s: cannot load options: %v\n%s", l.FileName, err, options)
	}
	return nil
}

func (l *Lesson) parseConfig(mdt *mdtext.MDText) error {
	l.ConfigSrc = mdt.Code["config"]
	if l.ConfigSrc == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(l.ConfigSrc), &(l.Config)); err != nil {
		return fmt.Errorf("%s: cannot load config: %v\n%s", l.FileName, err, l.ConfigSrc)
	}
	return nil
}

func FindLesson(allChapters map[string][]*Chapter, pathPrefix string, chapID, lessonID int) *Lesson {
	chapters := allChapters[pathPrefix]
	if chapters == nil {
		chapters = allChapters[""]
	}
	chap := get(chapID, chapters)
	return get(lessonID-1, chap.Content)
}
