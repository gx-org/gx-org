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

import "github.com/gx-org/gx-org/internal/lessons"

type (
	Page interface {
		DisplayLesson(*lessons.Lesson)
	}

	navActions struct {
		page    Page
		current *lessons.Lesson
	}
)

func (na *navActions) hasPrev() bool {
	if na.current == nil {
		return false
	}
	return na.current.Prev != nil
}

func (na *navActions) hasNext() bool {
	if na.current == nil {
		return false
	}
	return na.current.Next != nil
}

func (na *navActions) displayPrevious() {
	if !na.hasPrev() {
		return
	}
	na.page.DisplayLesson(na.current.Prev)
}

func (na *navActions) displayNext() {
	if !na.hasNext() {
		return
	}
	na.page.DisplayLesson(na.current.Next)
}
