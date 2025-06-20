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
		titleHTML string
		Content   []*Lesson
		ID        int
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

func New() ([]*Chapter, error) {
	var chapters []*Chapter
	chapterFound := true
	var prev *Lesson
	for chapterFound {
		chap := &Chapter{ID: len(chapters) + 1}
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
		chapters = append(chapters, chap)
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no content found")
	}
	return chapters, nil
}

func readLesson(chap *Chapter) (*Lesson, error) {
	lessonID := len(chap.Content) + 1
	fileName := fmt.Sprintf("%d_%d.md", chap.ID, lessonID)
	data, err := lessons.Lessons.ReadFile(fileName)
	if errors.Is(err, os.ErrNotExist) && (chap.ID > 1 || lessonID > 1) {
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

func (chap *Chapter) NumLessons() int {
	return len(chap.Content)
}

func FindLesson(chapters []*Chapter, chapID, lessonID int) *Lesson {
	chapI := chapID - 1
	lessonI := lessonID - 1
	if chapI <= 0 {
		return chapters[0].Content[0]
	}
	if chapI >= len(chapters) {
		return chapters[0].Content[0]
	}
	chap := chapters[chapI]
	if lessonI <= 0 {
		return chap.Content[0]
	}
	if lessonI >= len(chap.Content) {
		return chap.Content[0]
	}
	return chap.Content[lessonI]
}
