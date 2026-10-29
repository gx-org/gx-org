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

package lessons_test

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gx-org/gx-org/lessons"
	"github.com/gx-org/gx/api/gxmain"
	"github.com/gx-org/gx/build/builder"
	"github.com/gx-org/gx/golang/backend"
	"github.com/gx-org/gx/src/gx/format"
	"github.com/gx-org/gx/stdlib"
	"github.com/pkg/errors"
)

// skip lists the lessons that cannot be run, with the reason why.
var skip = map[string]string{
	"devs/demo_1.gx": "no Main function",
}

// run runs the Main function of a GX file like the gx run command does
// and returns what the command would print on the standard output.
func run(name string) (string, error) {
	vfs, err := fs.Sub(lessons.Lessons, path.Dir(name))
	if err != nil {
		return "", err
	}
	rtm, err := backend.New(builder.New(stdlib.Importer()))
	if err != nil {
		return "", err
	}
	fn, outs, traces, err := gxmain.FromFile(rtm, vfs, []string{path.Base(name)})
	if err != nil {
		return "", err
	}
	var out strings.Builder
	if err := gxmain.Printf(&out, fn, outs, traces); err != nil {
		return "", err
	}
	return out.String(), nil
}

// runLesson runs a GX lesson and compares its output with the expected
// output stored in <name>.out.
func runLesson(name string) error {
	want, err := fs.ReadFile(lessons.Lessons, name+".out")
	if err != nil {
		return err
	}
	got, err := run(name)
	if err != nil {
		return err
	}
	if diff := cmp.Diff(string(want), got); diff != "" {
		return errors.Errorf("unexpected output (-want +got):\n%s", diff)
	}
	return nil
}

// checkFormat checks that the source code of a GX lesson is formatted.
func checkFormat(name string) error {
	src, err := fs.ReadFile(lessons.Lessons, name)
	if err != nil {
		return err
	}
	want, err := format.Source(src)
	if err != nil {
		return err
	}
	if diff := cmp.Diff(string(want), string(src)); diff != "" {
		return errors.Errorf("source code is not formatted (-want +got):\n%s", diff)
	}
	return nil
}

func checkLesson(t *testing.T, name string) {
	if reason, ok := skip[name]; ok {
		t.Skip(reason)
	}
	if err := checkFormat(name); err != nil {
		t.Error(err)
	}
	if err := runLesson(name); err != nil {
		t.Error(err)
	}
}

func TestLessons(t *testing.T) {
	err := fs.WalkDir(lessons.Lessons, ".", func(name string, _ fs.DirEntry, err error) error {
		if err != nil || path.Ext(name) != ".gx" {
			return err
		}
		t.Run(name, func(t *testing.T) {
			checkLesson(t, name)
		})
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
