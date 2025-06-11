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

package ui

import "honnef.co/go/js/dom/v2"

type (
	ElementOption interface {
		Apply(dom.Element)
	}

	ElementOptionF func(dom.Element)
)

func (f ElementOptionF) Apply(el dom.Element) {
	f(el)
}

func Class(class string) ElementOption {
	return ElementOptionF(func(el dom.Element) {
		el.Class().Add(class)
	})
}

func Property(property, value string) ElementOption {
	return ElementOptionF(func(el dom.Element) {
		el.SetAttribute(property, value)
	})
}

func Listener(typ string, listener func(dom.Event)) ElementOption {
	return ElementOptionF(func(el dom.Element) {
		el.AddEventListener(typ, true, listener)
	})
}

func InnerHTML(s string) ElementOption {
	return ElementOptionF(func(el dom.Element) {
		el.SetInnerHTML(s)
	})
}

func SetVisible(visible bool) ElementOption {
	return ElementOptionF(func(el dom.Element) {
		propertyValue := ""
		if !visible {
			propertyValue = "none"
		}
		el.(dom.HTMLElement).Style().SetProperty("display", propertyValue, "")
	})
}

func applyAll(el dom.Element, opts []ElementOption) {
	for _, opt := range opts {
		opt.Apply(el)
	}
}
