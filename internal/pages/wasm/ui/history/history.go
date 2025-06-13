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

package history

type History[T any] struct {
	timeline []T
	next     int
}

func New[T any]() *History[T] {
	return &History[T]{}
}

func (h *History[T]) Append(t T) {
	h.timeline = h.timeline[:h.next]
	h.timeline = append(h.timeline, t)
	h.next = len(h.timeline)
}

func (h *History[T]) Undo() {
	if h.next == 0 {
		return
	}
	h.next--
}

func (h *History[T]) Redo() {
	if h.next >= len(h.timeline) {
		return
	}
	h.next++
}

func (h *History[T]) History() []T {
	return h.timeline
}

func (h *History[T]) Current() T {
	var current T
	if h.next == 0 {
		return current
	}
	return h.timeline[h.next-1]
}
