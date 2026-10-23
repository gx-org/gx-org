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

//go:build js && wasm

package ace

import (
	"syscall/js"
)

type Session struct {
	js.Value
}

func (ed *Editor) Session() *Session {
	return &Session{
		Value: ed.Get("session"),
	}
}

func (se *Session) SetMode(mode string) {
	se.Call("setMode", mode)
}

func (se *Session) OnChange(f func()) {
	se.Call("on", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		f()
		return nil
	}))
}

// Annotation is a message displayed in the gutter of the editor.
type Annotation struct {
	// Row and Column of the annotation (zero-based).
	Row, Column int
	// Text displayed when hovering over the annotation.
	Text string
	// Type of the annotation: "error", "warning", or "info".
	Type string
}

// SetAnnotations replaces all the annotations of the session.
func (se *Session) SetAnnotations(anns []Annotation) {
	jsAnns := make([]any, len(anns))
	for i, ann := range anns {
		jsAnns[i] = map[string]any{
			"row":    ann.Row,
			"column": ann.Column,
			"text":   ann.Text,
			"type":   ann.Type,
		}
	}
	se.Call("setAnnotations", jsAnns)
}

// Range of text in the session.
type Range struct {
	// StartRow and StartColumn of the range (zero-based).
	StartRow, StartColumn int
	// EndRow and EndColumn of the range (zero-based, excluded).
	EndRow, EndColumn int
}

// jsValue returns the range as an ACE Range javascript object.
func (r Range) jsValue() js.Value {
	rangeClass := js.Global().Get("ace").Call("require", "ace/range").Get("Range")
	return rangeClass.New(r.StartRow, r.StartColumn, r.EndRow, r.EndColumn)
}

// AddMarker highlights a range of text with a CSS class.
// The marker is drawn in front of the text.
// It returns the ID of the marker to remove it with RemoveMarker.
func (se *Session) AddMarker(rng Range, class string) int {
	return se.Call("addMarker", rng.jsValue(), class, "text", true).Int()
}

// RemoveMarker removes a marker added with AddMarker.
func (se *Session) RemoveMarker(id int) {
	se.Call("removeMarker", id)
}
