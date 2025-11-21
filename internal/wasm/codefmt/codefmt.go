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

package codefmt

import (
	"fmt"
	"html"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type Formatter struct {
	lexer chroma.Lexer
	fmt   *chromahtml.Formatter
	style *chroma.Style
}

func Go() *Formatter {
	return newFormatter(lexers.Go)
}

func JSON() *Formatter {
	return newFormatter(lexers.Get("JSON"))
}

func newFormatter(lexer chroma.Lexer) *Formatter {
	return &Formatter{
		lexer: lexer,
		fmt: chromahtml.New(
			chromahtml.Standalone(false),
			chromahtml.WithClasses(true),
			chromahtml.ClassPrefix("chroma_"),
		),
		style: styles.Get("xcode"),
	}
}

func defaultFormat(s string) string {
	s = strings.ReplaceAll(s, "\t", "    ")
	s = strings.ReplaceAll(s, " ", "\u00a0")
	s = html.EscapeString(s)
	return s
}

func (ft Formatter) Format(src string) string {
	it, err := ft.lexer.Tokenise(nil, src)
	if err != nil {
		fmt.Printf("cannot tokenise source code: %v\nSource:\n%s", err.Error(), src)
		return defaultFormat(src)
	}
	var w strings.Builder
	if err := ft.fmt.Format(&w, ft.style, it); err != nil {
		fmt.Printf("cannot tokenise source code: %v\nSource:\n%s", err.Error(), src)
		return defaultFormat(src)
	}
	return w.String()
}
