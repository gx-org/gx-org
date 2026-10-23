// Copyright 2026 Google LLC
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

package editor

import (
	"strings"
	"testing"
	"unicode/utf16"
)

// cursorMark marks the position of the cursor in the tests.
const cursorMark = "‸"

// splitCursor removes the cursor mark from s and returns the position of the cursor
// in UTF-16 code units.
func splitCursor(t *testing.T, s string) (string, int) {
	t.Helper()
	before, after, found := strings.Cut(s, cursorMark)
	if !found {
		t.Fatalf("no cursor in %q", s)
	}
	return before + after, len(utf16.Encode([]rune(before)))
}

func TestReformattedCursor(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		formatted string
	}{
		{
			name:      "end of a line",
			text:      "\tx:=1‸\n\treturn x\n",
			formatted: "\tx := 1‸\n\treturn x\n",
		},
		{
			name:      "before some code",
			text:      "\tx:=‸1\n",
			formatted: "\tx := ‸1\n",
		},
		{
			name:      "after a removed semicolon",
			text:      "\tx := 1;‸\n",
			formatted: "\tx := 1‸\n",
		},
		{
			name:      "split line",
			text:      "\tx := 1; ‸y := 2\n",
			formatted: "\tx := 1\n\t‸y := 2\n",
		},
		{
			name:      "after a split line",
			text:      "\tfor _, ax := range shape { n *= ax }\n\treturn n‸\n",
			formatted: "\tfor _, ax := range shape {\n\t\tn *= ax\n\t}\n\treturn n‸\n",
		},
		{
			name:      "indentation",
			text:      "{\n    ‸x := 1\n}\n",
			formatted: "{\n\t‸x := 1\n}\n",
		},
		{
			name:      "start of an indented line",
			text:      "{\n‸    x := 1\n}\n",
			formatted: "{\n‸\tx := 1\n}\n",
		},
		{
			name:      "empty line",
			text:      "\tx := 1\n\t‸\n\treturn x\n",
			formatted: "\tx := 1\n‸\n\treturn x\n",
		},
		{
			name:      "removed empty lines",
			text:      "\tx := 1\n\n\n‸\n\treturn x\n",
			formatted: "\tx := 1\n\n‸\treturn x\n",
		},
		{
			name:      "in white spaces",
			text:      "\treturn 1 ‸ +  2\n",
			formatted: "\treturn 1 ‸+ 2\n",
		},
		{
			name:      "trailing white spaces",
			text:      "\treturn 1  ‸\n",
			formatted: "\treturn 1‸\n",
		},
		{
			name:      "UTF-16 surrogate pairs",
			text:      "// 😀 é\n\treturn 1+2‸\n",
			formatted: "// 😀 é\n\treturn 1 + 2‸\n",
		},
		{
			name:      "beginning of the text",
			text:      "‸package main",
			formatted: "‸package main\n",
		},
		{
			name:      "end of the text",
			text:      "}\n\n\n‸",
			formatted: "}\n‸",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, cursor := splitCursor(t, test.text)
			formatted, want := splitCursor(t, test.formatted)
			got := reformattedCursor(text, formatted, cursor)
			if got != want {
				gotOffset := byteOffset(formatted, got)
				t.Errorf("incorrect cursor position: got %d (%q) but want %d (%q)", got, formatted[:gotOffset]+cursorMark+formatted[gotOffset:], want, test.formatted)
			}
		})
	}
}
