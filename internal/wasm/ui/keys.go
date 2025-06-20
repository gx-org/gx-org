package ui

import (
	"fmt"

	"honnef.co/go/js/dom/v2"
)

// Keys maintain a map of keys being pressed.
type Keys struct {
	downs map[string]bool
	cb    func(*Keys, *dom.KeyboardEvent)
}

func KeyListener(cb func(*Keys, *dom.KeyboardEvent)) ElementOption {
	keys := &Keys{downs: make(map[string]bool), cb: cb}
	keyDown := Listener("keydown", keys.onKeyDown)
	keyUp := Listener("keyup", keys.onKeyUp)
	return ElementOptionF(func(el dom.Element) {
		keyDown.Apply(el)
		keyUp.Apply(el)
	})
}

func (k *Keys) onKeyUp(ev *dom.KeyboardEvent) {
	delete(k.downs, ev.Key())
}

var noEmulation = map[string]bool{
	"Meta":  true,
	"Shift": true,
}

func (k *Keys) onKeyDown(ev *dom.KeyboardEvent) {
	k.downs[ev.Key()] = true
	k.cb(k, ev)
	if k.On("Meta") && !noEmulation[ev.Key()] {
		// On MacOS, we do not receive a key down event when
		// the meta key is pressed. So, we emulate it.
		k.onKeyUp(ev)
	}
}

func (k *Keys) On(keys ...string) bool {
	for _, key := range keys {
		if !k.downs[key] {
			return false
		}
	}
	return true
}

func (k *Keys) String() string {
	return fmt.Sprint(k.downs)
}
