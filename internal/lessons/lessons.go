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
	"errors"
	"fmt"
	"os"

	"github.com/gx-org/gx-org/internal/mdtext"
	"github.com/gx-org/gx-org/lessons"
)

type (
	Chapter struct {
		ID        int
		name      string
		titleHTML string

		Content []*Lesson
	}

	Lesson struct {
		Chapter *Chapter
		ID      int

		HTML string
		Code string

		Prev *Lesson
		Next *Lesson
	}
)

var chapterNames = []string{
	"intro",
	"types",
}

func New() (map[string]*Chapter, error) {
	chapters := make(map[string]*Chapter, len(chapterNames))
	var prev *Lesson
	for _, name := range chapterNames {
		chap := &Chapter{name: name, ID: len(chapters) + 1}
		chapters[name] = chap
		lessonFound := true
		for lessonFound {
			lesson, err := readLesson(chap)
			if err != nil {
				return nil, err
			}
			if lesson != nil {
				lessonFound = true
				if prev != nil {
					lesson.Prev = prev
					prev.Next = lesson
				}
				prev = lesson
			} else {
				lessonFound = false
			}
		}
		if len(chap.Content) == 0 {
			break
		}
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no content found")
	}
	return chapters, nil
}

func readLesson(chap *Chapter) (*Lesson, error) {
	lessonID := len(chap.Content) + 1
	fileName := fmt.Sprintf("%s_%d.md", chap.Name(), lessonID)
	data, err := lessons.Lessons.ReadFile(fileName)
	if errors.Is(err, os.ErrNotExist) && lessonID > 1 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", fileName, err)
	}
	mdt := mdtext.Parse(data)
	lesson := &Lesson{Chapter: chap, ID: lessonID}
	if mdt.TitleHTML != "" && lessonID != 1 {
		return nil, fmt.Errorf("%s: chapter title can only be specified for the first lesson", fileName)
	}
	if mdt.TitleHTML == "" && lessonID == 1 {
		return nil, fmt.Errorf("%s: no chapter title specified", fileName)
	}
	if lessonID == 1 {
		chap.titleHTML = mdt.TitleHTML
	}
	lesson.HTML = chap.titleHTML + "\n\n" + mdt.HTML
	lesson.Code = mdt.Code[mdtext.TagPrefix+"code"]
	if lesson.Code == "" {
		return nil, fmt.Errorf("lesson %s has no GX source code", fileName)
	}
	chap.Content = append(chap.Content, lesson)
	return lesson, nil
}

func (chap *Chapter) Name() string {
	return chap.name
}

func (chap *Chapter) NumLessons() int {
	return len(chap.Content)
}

func FindLesson(chapters map[string]*Chapter, chapName string, lessonID int) *Lesson {
	if chapName == "" {
		chapName = chapterNames[0]
	}
	lessonI := lessonID - 1
	chap, ok := chapters[chapName]
	if !ok {
		return nil
	}
	if lessonI <= 0 {
		return chap.Content[0]
	}
	if lessonI >= len(chap.Content) {
		return chap.Content[0]
	}
	return chap.Content[lessonI]
}
