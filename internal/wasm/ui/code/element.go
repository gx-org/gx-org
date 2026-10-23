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

package code

import (
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"runtime/debug"
	"strings"
	"unicode/utf16"

	"github.com/gomlx/compute/gobackend"
	"github.com/gx-org/gx-org/internal/lessons"
	"github.com/gx-org/gx-org/internal/wasm/ace"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"github.com/gx-org/gx/api"
	"github.com/gx-org/gx/api/hostio"
	"github.com/gx-org/gx/api/runner"
	"github.com/gx-org/gx/api/values"
	"github.com/gx-org/gx/build/builder"
	"github.com/gx-org/gx/build/fmterr"
	"github.com/gx-org/gx/build/ir"
	"github.com/gx-org/gx/golang/backend/kernels"
	"github.com/gx-org/gx/stdlib"
	"honnef.co/go/js/dom/v2"
)

type Code struct {
	gui *ui.UI
	src *Source
	out *Output

	bld *builder.Builder
}

func New(gui *ui.UI, parent dom.HTMLElement) *Code {
	bld := builder.New(stdlib.Importer())
	cd := &Code{
		gui: gui,
		bld: bld,
	}
	container := gui.CreateDIV(parent, ui.Class("code_container"))
	cd.src = newSource(cd, container)
	cd.out = newOutput(cd, container)

	return cd
}

func (cd *Code) SetContent(les *lessons.Lesson) {
	cd.src.editor.Set(les.Code)
}

func (cd *Code) compile(src string) error {
	_, err := cd.compileCode(src)
	if err != nil {
		return err
	}
	cd.out.setHTML("")
	return nil
}

func (cd *Code) compileCode(src string) (*ir.Package, error) {
	pkg := cd.bld.NewIncrementalPackage("main")
	err := pkg.Build(src)
	anns, rngs := errorAnnotations(src, err)
	cd.src.editor.SetAnnotations(anns)
	cd.src.editor.SetErrorMarkers(rngs)
	if err != nil {
		return nil, err
	}
	return pkg.IR(), nil
}

// errorAnnotations converts compiler errors into annotations for the code editor
// and into ranges of code to underline with a red squiggly line.
func errorAnnotations(src string, err error) (anns []ace.Annotation, rngs []ace.Range) {
	var errs *fmterr.Errors
	if !errors.As(err, &errs) {
		return nil, nil
	}
	for _, err := range errs.Errors() {
		var posErr fmterr.ErrorWithPos
		if !errors.As(err, &posErr) {
			continue
		}
		pos := posErr.Pos().Begin
		if !pos.IsValid() || pos.Offset > len(src) {
			continue
		}
		row, column := acePosition(src, pos.Offset)
		anns = append(anns, ace.Annotation{
			Row:    row,
			Column: column,
			Text:   posErr.Err().Error(),
			Type:   "error",
		})
		start, end := errorToken(src, pos.Offset)
		rngs = append(rngs, aceRange(src, start, end))
	}
	return anns, rngs
}

// errorToken returns the byte offsets [start, end) in src of the token
// where an error at offset is. If there is no token at offset (for example,
// the error is at the end of a line), it returns the previous token on the
// same line, or an empty range if there is none.
func errorToken(src string, offset int) (start, end int) {
	lineStart := strings.LastIndexByte(src[:offset], '\n') + 1
	lineEnd := len(src)
	if i := strings.IndexByte(src[offset:], '\n'); i >= 0 {
		lineEnd = offset + i
	}
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var sc scanner.Scanner
	sc.Init(file, []byte(src), nil, scanner.ScanComments)
	start, end = offset, offset
	for {
		pos, tok, lit := sc.Scan()
		tokStart := file.Offset(pos)
		if tok == token.EOF || tokStart > offset {
			return start, end
		}
		if tok == token.SEMICOLON && lit == "\n" {
			continue // Semicolon automatically inserted at the end of a line.
		}
		tokEnd := file.Offset(sc.End())
		if tokEnd > offset {
			// The token contains the offset: keep it on the line of the error.
			return max(tokStart, lineStart), min(tokEnd, lineEnd)
		}
		if tok != token.COMMENT && tokStart >= lineStart {
			start, end = tokStart, tokEnd
		}
	}
}

// aceRange converts the byte offsets [start, end) in src into a range in the code editor.
func aceRange(src string, start, end int) ace.Range {
	startRow, startColumn := acePosition(src, start)
	endRow, endColumn := acePosition(src, end)
	return ace.Range{
		StartRow:    startRow,
		StartColumn: startColumn,
		EndRow:      endRow,
		EndColumn:   endColumn,
	}
}

// acePosition converts a byte offset in src into a row and a column in the code editor.
// Columns are counted in UTF-16 code units, like Javascript strings.
func acePosition(src string, offset int) (row, column int) {
	before := src[:offset]
	for _, r := range before[strings.LastIndexByte(before, '\n')+1:] {
		column += utf16.RuneLen(r)
	}
	return strings.Count(before, "\n"), column
}

func (cd *Code) updateCodeOutput(f func(src string) error, src string) {
	defer func() {
		if r := recover(); r != nil {
			cd.out.setHTML(toHTML(fmt.Errorf("GX PANIC: please report everything below so that it can be fixed:\n%s\n%s", src, debug.Stack())))
		}
	}()
	if err := f(src); err != nil {
		cd.out.setHTML(toHTML(err.Error()))
		return
	}
}

func flatten(out []hostio.Value) []hostio.Value {
	flat := []hostio.Value{}
	for _, v := range out {
		slice, ok := v.(*values.Slice)
		if !ok {
			flat = append(flat, v)
			continue
		}
		vals := make([]hostio.Value, slice.Len())
		for i := 0; i < slice.Len(); i++ {
			vals[i] = slice.Element(i)
		}
		flat = append(flat, flatten(vals)...)
	}
	return flat
}

func buildString[T any](w *strings.Builder, vals []T) {
	if len(vals) == 0 {
		return
	}
	htmls := make([]string, len(vals))
	for i, val := range vals {
		htmls[i] = toHTML(val)
	}
	if len(htmls) == 1 {
		w.WriteString(htmls[0])
		return
	}
	enumerateB(w, htmls)
}

type traceWriter struct {
	buf strings.Builder
}

func (r *traceWriter) Trace(file *ir.File, call *ir.FuncCallExpr, vals []hostio.Value) error {
	vals, err := hostio.ToHost(kernels.DefaultAllocator(), vals)
	if err != nil {
		return err
	}
	for _, val := range vals {
		toHTMLB(&r.buf, val)
	}
	return nil
}

func (cd *Code) runFuncCompEval(fun *ir.FuncDecl) string {
	_, outs, err := builder.CompEvalFunc(cd.bld, fun)
	if err != nil {
		return err.Error()
	}
	bld := strings.Builder{}
	buildString(&bld, outs)
	return bld.String()
}

func (cd *Code) runFuncDecl(fun *ir.FuncDecl) string {
	numArgs := fun.FuncType().Params.Len()
	if numArgs > 0 {
		return errorF("func Main must have no arguments")
	}
	bck, err := gobackend.New("")
	if err != nil {
		return errorF("cannot create Go native backend: %v", err)
	}
	dev, err := api.NewRuntime(bck, cd.bld).Device(0)
	if err != nil {
		return errorF("cannot create device: %v", err)
	}
	runner, err := runner.Trace(dev, fun, nil, nil, nil)
	if err != nil {
		return errorF("cannot compile the code: %v", err)
	}
	var trace traceWriter
	vals, err := runner.Run(nil, nil, &trace)
	if err != nil {
		return err.Error()
	}
	flattenVals, err := hostio.ToHost(kernels.DefaultAllocator(), flatten(vals))
	if err != nil {
		return err.Error()
	}
	bld := strings.Builder{}
	buildString(&bld, flattenVals)
	if trace.buf.Len() > 0 {
		bld.WriteString(fmt.Sprintf("Trace:<br>%s", trace.buf.String()))
	}
	return bld.String()
}

func (cd *Code) runFunc(fun ir.PkgFunc) string {
	if fun == nil {
		return errorF("Main function not found")
	}
	funDecl, isFuncDecl := fun.(*ir.FuncDecl)
	if !isFuncDecl {
		return "cannot run builtin functions"
	}
	if fun.FuncType().Nature.CompEval {
		return cd.runFuncCompEval(funDecl)
	}
	return cd.runFuncDecl(funDecl)
}

func (cd *Code) compileAndRun(src string) error {
	irPkg, err := cd.compileCode(src)
	if err != nil {
		return err
	}
	out := cd.runFunc(irPkg.FindFunc("Main"))
	cd.out.setHTML(out)
	return nil
}
