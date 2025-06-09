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
	}
)

func read(archFS fs.FS, name string) (string, error) {
	data, err := fs.ReadFile(archFS, name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readLesson(chap *Chapter) (bool, error) {
	lessonID := len(chap.Content) + 1
	fileName := fmt.Sprintf("lesson_%d_%d.txtar", chap.ID, lessonID)
	data, err := lessons.Lessons.ReadFile(fileName)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cannot read %s: %v", fileName, err)
	}
	arch := txtar.Parse(data)
	archFS, err := txtar.FS(arch)
	if err != nil {
		return false, err
	}
	lesson := &Lesson{Chapter: chap, ID: lessonID}
	lesson.Text, err = read(archFS, "lesson")
	if err != nil {
		return false, fmt.Errorf("cannot read file lesson in archive %s: %v", fileName, err)
	}
	lesson.Code, err = read(archFS, "code")
	if err != nil {
		return false, fmt.Errorf("cannot read file code in archive %s: %v", fileName, err)
	}
	chap.Content = append(chap.Content, lesson)
	return true, nil
}

func New() ([]*Chapter, error) {
	var chapters []*Chapter
	chapterFound := true
	for chapterFound {
		chap := &Chapter{ID: len(chapters) + 1}
		lessonFound := true
		for lessonFound {
			var err error
			lessonFound, err = readLesson(chap)
			if err != nil {
				return nil, err
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
