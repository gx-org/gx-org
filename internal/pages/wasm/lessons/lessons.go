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
	"io/fs"
	"os"
	"strings"

	"github.com/gx-org/gx-org/lessons"
	"golang.org/x/tools/txtar"
)

type (
	Chapter struct {
		Title   string
		Content []*Lesson
		ID      int
	}

	Lesson struct {
		Chapter *Chapter
		ID      int

		Text string
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

func read(archFS fs.FS, name string) (string, error) {
	data, err := fs.ReadFile(archFS, name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readTitle(s string) (string, error) {
	var title string
	for line := range strings.Lines(s) {
		splits := strings.SplitN(line, ":", 2)
		if len(splits) != 2 {
			continue
		}
		if splits[0] != "Chapter" {
			continue
		}
		if title != "" {
			return "", fmt.Errorf("cannot specified more than one title")
		}
		title = splits[1]
	}
	return title, nil
}

func readLesson(chap *Chapter) (*Lesson, error) {
	lessonID := len(chap.Content) + 1
	fileName := fmt.Sprintf("lesson_%d_%d.txtar", chap.ID, lessonID)
	data, err := lessons.Lessons.ReadFile(fileName)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %v", fileName, err)
	}
	arch := txtar.Parse(data)
	archFS, err := txtar.FS(arch)
	if err != nil {
		return nil, err
	}
	lesson := &Lesson{Chapter: chap, ID: lessonID}
	title, err := readTitle(string(arch.Comment))
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", fileName, err)
	}
	if title != "" && lessonID != 1 {
		return nil, fmt.Errorf("%s: chapter title can only be specified for the first lesson", fileName)
	}
	if title == "" && lessonID == 1 {
		return nil, fmt.Errorf("%s: no chapter title specified", fileName)
	}
	if lessonID == 1 {
		chap.Title = title
	}
	lesson.Text, err = read(archFS, "lesson")
	if err != nil {
		return nil, fmt.Errorf("cannot read file lesson in archive %s: %v", fileName, err)
	}
	lesson.Code, err = read(archFS, "code")
	if err != nil {
		return nil, fmt.Errorf("cannot read file code in archive %s: %v", fileName, err)
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
