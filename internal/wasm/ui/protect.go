package ui

import (
	"fmt"
	"runtime/debug"
)

var crashCallback func(string) = printToConsole

func printToConsole(crash string) {
	fmt.Println(crash)
}

func SetCrashOutput(f func(string)) {
	if f == nil {
		f = printToConsole
	}
	crashCallback = f
}

func Protect(f func()) {
	defer func() {
		if r := recover(); r != nil {
			crashCallback(fmt.Sprintf("UI crash: %s.\nPlease report the following stackstrace:\n%s\n", r, string(debug.Stack())))
		}
	}()
	f()
}

func Go(f func()) {
	go Protect(f)
}
