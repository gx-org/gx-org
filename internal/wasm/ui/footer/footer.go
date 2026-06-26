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

package footer

import (
	"fmt"
	"runtime/debug"

	"github.com/gx-org/gx-org/internal/lessons"
	"github.com/gx-org/gx-org/internal/wasm/buildtime"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Footer struct {
	el dom.HTMLElement
}

func New(gui *ui.UI) *Footer {
	els := gui.Dom().Document().GetElementsByTagName("footer")
	hdr := &Footer{}
	if len(els) == 0 {
		return hdr
	}
	hdr.el = els[0].(dom.HTMLElement)
	hdr.SetContent(nil)
	return hdr
}

func gxVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "no build info"
	}
	for _, dep := range buildInfo.Deps {
		if dep.Path != "github.com/gx-org/gx" {
			continue
		}
		return dep.Version
	}
	return "no version"
}

func (hdr *Footer) SetContent(les *lessons.Lesson) {
	hdr.el.SetInnerHTML(fmt.Sprintf("Build at %s GX: %s", buildtime.BuildTime, gxVersion()))
}
