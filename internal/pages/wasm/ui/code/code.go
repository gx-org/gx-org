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

package code

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gx-org/gx-org/internal/pages/wasm/lessons"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui"
	"github.com/gx-org/gx/api"
	"github.com/gx-org/gx/api/tracer"
	"github.com/gx-org/gx/api/values"
	"github.com/gx-org/gx/build/builder"
	"github.com/gx-org/gx/build/importers"
	"github.com/gx-org/gx/build/ir"
	"github.com/gx-org/gx/golang/backend"
	"github.com/gx-org/gx/golang/backend/kernels"
	"github.com/gx-org/gx/stdlib"
	"honnef.co/go/js/dom/v2"
)

type Code struct {
	gui *ui.UI
	src *Source
	out *Output

	bld    *builder.Builder
	dev    *api.Device
	devErr error
}

func New(gui *ui.UI, parent dom.HTMLElement) *Code {
	bld := builder.New(importers.NewCacheLoader(
		stdlib.Importer(nil),
	))
	cd := &Code{
		gui: gui,
		bld: bld,
	}
	container := gui.CreateDIV(parent, ui.Class("code_container"))
	cd.src = newSource(cd, container)
	cd.out = newOutput(cd, container)

	cd.dev, cd.devErr = backend.New(bld).Device(0)
	return cd
}

func (cd *Code) SetContent(les *lessons.Lesson) {
	cd.src.set(les.Code, nil)
}

func (cd *Code) compileAndWrite(src string) error {
	_, err := cd.compileCode(src)
	if err != nil {
		return err
	}
	cd.out.set("")
	return nil
}

func (cd *Code) compileCode(src string) (*ir.Package, error) {
	if cd.devErr != nil {
		return nil, fmt.Errorf("Cannot initialise backend: %s", cd.devErr.Error())
	}
	pkg := cd.bld.NewIncrementalPackage("main")
	if err := pkg.Build(src); err != nil {
		return nil, err
	}
	return pkg.IR(), nil
}

func (cd *Code) callAndWrite(f func(src string) error, src string) {
	defer func() {
		if r := recover(); r != nil {
			cd.out.set(fmt.Sprintf("GX PANIC: please report everything below so that it can be fixed:\n%s\n%s", src, debug.Stack()))
		}
	}()
	if err := f(src); err != nil {
		cd.out.set(fmt.Sprintf("ERROR: %s", err.Error()))
		return
	}
}

func flatten(out []values.Value) []values.Value {
	flat := []values.Value{}
	for _, v := range out {
		slice, ok := v.(*values.Slice)
		if !ok {
			flat = append(flat, v)
			continue
		}
		vals := make([]values.Value, slice.Size())
		for i := 0; i < slice.Size(); i++ {
			vals[i] = slice.Element(i)
		}
		flat = append(flat, flatten(vals)...)
	}
	return flat
}

func buildString(bld *strings.Builder, out []values.Value) error {
	out, err := values.ToHost(kernels.Allocator(), flatten(out))
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return nil
	}
	if len(out) == 1 {
		bld.WriteString(fmt.Sprint(out[0]))
		return nil
	}
	for i, s := range out {
		bld.WriteString(fmt.Sprintf("%d: %v\n", i, s))
	}
	return nil
}

func (cd *Code) runFunc(fun ir.Func, args []values.Value) ([]values.Value, string, bool) {
	numArgs := fun.FuncType().Params.Len()
	if len(args) < numArgs {
		return nil, fmt.Sprintf("not enough arguments to pass to %s: got %d but want %d", fun.Name(), len(args), numArgs), false
	}
	args = args[:numArgs]
	runner, err := tracer.Trace(cd.dev, fun.(*ir.FuncDecl), nil, args, nil)
	if err != nil {
		return nil, err.Error(), false
	}
	vals, err := runner.Run(nil, args, nil)
	if err != nil {
		return nil, err.Error(), false
	}
	bld := strings.Builder{}
	if err := buildString(&bld, vals); err != nil {
		return nil, err.Error(), false
	}
	return vals, bld.String(), true
}

func indent(s string) string {
	var lines []string
	for line := range strings.Lines(s) {
		lines = append(lines, "  "+line)
	}
	if lines[len(lines)-1] != "\n" {
		lines = append(lines, "\n")
	}
	return strings.Join(lines, "")
}

func (cd *Code) runCode(src string) error {
	irPkg, err := cd.compileCode(src)
	if err != nil {
		return err
	}
	bld := strings.Builder{}
	var vals []values.Value
	for fun := range irPkg.ExportedFuncs() {
		bld.WriteString(fun.Name() + ":\n")
		var s string
		var ok bool
		vals, s, ok = cd.runFunc(fun, vals)
		bld.WriteString(indent(s))
		if !ok {
			break
		}
	}
	cd.out.set(bld.String())
	return nil
}
