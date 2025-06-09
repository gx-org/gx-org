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

	"github.com/gx-org/gx/api"
	"github.com/gx-org/gx/api/tracer"
	"github.com/gx-org/gx/api/values"
	"github.com/gx-org/gx/build/builder"
	"github.com/gx-org/gx/build/importers"
	"github.com/gx-org/gx/build/ir"
	"github.com/gx-org/gx/golang/backend"
	"github.com/gx-org/gx/golang/backend/kernels"
	"github.com/gx-org/gx/stdlib"
	"github.com/tdegris/tdegris/internal/pages/wasm/lessons"
	"github.com/tdegris/tdegris/internal/pages/wasm/ui"
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
	cd.src.set(les.Code)
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

func buildString(out []values.Value) (string, error) {
	out, err := values.ToHost(kernels.Allocator(), flatten(out))
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", nil
	}
	if len(out) == 1 {
		return fmt.Sprint(out[0]), nil
	}
	bld := strings.Builder{}
	for i, s := range out {
		bld.WriteString(fmt.Sprintf("%d: %v\n", i, s))
	}
	return bld.String(), nil
}

func (cd *Code) runCode(src string) error {
	irPkg, err := cd.compileCode(src)
	if err != nil {
		return err
	}
	const fnName = "Main"
	fn := irPkg.FindFunc(fnName)
	if fn == nil {
		return fmt.Errorf("function %s not found", fnName)
	}
	runner, err := tracer.Trace(cd.dev, fn.(*ir.FuncDecl), nil, nil, nil)
	if err != nil {
		return err
	}
	values, err := runner.Run(nil, nil, nil)
	if err != nil {
		return err
	}
	outS, err := buildString(values)
	if err != nil {
		return err
	}
	cd.out.set(outS)
	return nil
}
