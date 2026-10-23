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
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// isCode returns true if a rune is kept when the code is formatted.
// Formatting the code only changes white spaces and removes semicolons.
func isCode(r rune) bool {
	return r != ';' && !unicode.IsSpace(r)
}

// reformattedCursor returns the position of the cursor in formatted, a formatted version of text,
// such that the cursor stays at the same place in the code.
// Positions count UTF-16 code units, like in a Javascript string.
func reformattedCursor(text, formatted string, cursor int) int {
	offset := reformattedOffset(text, formatted, byteOffset(text, cursor))
	return utf16Len(formatted[:offset])
}

// reformattedOffset returns the byte offset in formatted corresponding to the byte offset in text.
//
// If the cursor is just before some code, it is moved just before the same code in formatted.
// Otherwise, the cursor is after some white spaces (or semicolons) following some code:
// it is moved after the same code in formatted, then, as much as possible, to the same line
// and to the same column within the white spaces.
func reformattedOffset(text, formatted string, offset int) int {
	// Count the code runes before the cursor and find where they end.
	numCode, codeEnd := 0, 0
	for i := 0; i < offset; {
		r, size := utf8.DecodeRuneInString(text[i:offset])
		i += size
		if isCode(r) {
			numCode++
			codeEnd = i
		}
	}
	// Move after the same number of code runes in formatted.
	pos := 0
	for i := 0; numCode > 0 && i < len(formatted); {
		r, size := utf8.DecodeRuneInString(formatted[i:])
		i += size
		if isCode(r) {
			numCode--
			pos = i
		}
	}
	if offset < len(text) {
		if r, _ := utf8.DecodeRuneInString(text[offset:]); isCode(r) {
			// The cursor is just before some code: move it just before the next code rune.
			for pos < len(formatted) {
				r, size := utf8.DecodeRuneInString(formatted[pos:])
				if isCode(r) {
					break
				}
				pos += size
			}
			return pos
		}
	}
	// Move to the same line and the same column within the white spaces.
	spaces := text[codeEnd:offset]
	lines := strings.Count(spaces, "\n")
	column := utf8.RuneCountInString(spaces[strings.LastIndexByte(spaces, '\n')+1:])
	for i := pos; lines > 0 && i < len(formatted); {
		r, size := utf8.DecodeRuneInString(formatted[i:])
		if isCode(r) {
			break
		}
		i += size
		if r == '\n' {
			pos = i
			lines--
		}
	}
	for ; column > 0 && pos < len(formatted); column-- {
		r, size := utf8.DecodeRuneInString(formatted[pos:])
		if r == '\n' || isCode(r) {
			break
		}
		pos += size
	}
	return pos
}

// byteOffset converts a position counting UTF-16 code units in s into a byte offset.
func byteOffset(s string, pos int) int {
	for i, r := range s {
		if pos <= 0 {
			return i
		}
		pos -= utf16.RuneLen(r)
	}
	return len(s)
}

// utf16Len returns the number of UTF-16 code units encoding s.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}
