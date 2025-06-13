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

package history_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gx-org/gx-org/internal/pages/wasm/ui/history"
)

type checker struct {
	hist *history.History[int]
	num  int
}

func (c *checker) checkCurrent(t *testing.T, want int) {
	defer func() {
		c.num++
	}()
	got := c.hist.Current()
	if !cmp.Equal(got, want) {
		t.Errorf("check %d: got %v but want %v", c.num, got, want)
	}
}

func (c *checker) checkHistory(t *testing.T, want []int) {
	defer func() {
		c.num++
	}()
	got := c.hist.History()
	if !cmp.Equal(got, want) {
		t.Errorf("check %d: got %v but want %v", c.num, got, want)
	}
}

func TestHistory(t *testing.T) {
	hist := history.New[int]()
	c := &checker{hist: hist}
	for i := range 4 {
		hist.Append(i)
	}
	c.checkHistory(t, []int{0, 1, 2, 3})
	c.checkCurrent(t, 3)

	hist.Undo()
	hist.Undo()
	c.checkCurrent(t, 1)

	hist.Redo()
	c.checkCurrent(t, 2)
	hist.Redo()
	c.checkCurrent(t, 3)
	hist.Redo()
	c.checkCurrent(t, 3)

	for range 10 {
		hist.Undo()
	}
	c.checkCurrent(t, 0)

	for range 10 {
		hist.Redo()
	}
	c.checkCurrent(t, 3)

	hist.Undo()
	hist.Undo()
	c.checkCurrent(t, 1)

	hist.Append(100)
	c.checkCurrent(t, 100)
	c.checkHistory(t, []int{0, 1, 100})
}
