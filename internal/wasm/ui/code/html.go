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

package code

import (
	"fmt"
	"strings"

	"github.com/gx-org/gx-org/internal/wasm/codefmt"
	"github.com/gx-org/gx/api/values"
)

func preB(w *strings.Builder, x any) {
	fmt.Fprintf(w, `<pre class="chroma_chroma">%s</pre>`, x)
}

func errorB(w *strings.Builder, err error) {
	preB(w, err.Error())
}

func errorF(s string, a ...any) string {
	var bld strings.Builder
	errorB(&bld, fmt.Errorf(s, a...))
	return bld.String()
}

func enumerateB(w *strings.Builder, ss []string) {
	fmt.Fprintln(w, `<ol class="code_output_list">`)
	for _, s := range ss {
		fmt.Fprintf(w, "  <li>%s</li>\n", s)
	}
	fmt.Fprintln(w, "</ol>")
}

var formatter = codefmt.Go()

func irNodeHTML(w *strings.Builder, node string) {
	src := formatter.Format(node)
	w.WriteString(src)
}

func toHTMLB(w *strings.Builder, val any) {
	switch valT := val.(type) {
	case *values.IRNode:
		irNodeHTML(w, valT.String())
	case error:
		errorB(w, valT)
	default:
		preB(w, valT)
	}
}

func toHTML(val any) string {
	var bld strings.Builder
	toHTMLB(&bld, val)
	return bld.String()
}
