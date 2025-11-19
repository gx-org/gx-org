package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func mainErr() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: chromacss <target path>")
	}
	ft := html.New(
		html.Standalone(false),
		html.WithClasses(true),
		html.ClassPrefix("chroma_"),
	)
	var css strings.Builder
	if err := ft.WriteCSS(&css, styles.Get("xcode")); err != nil {
		return err
	}
	return os.WriteFile(os.Args[1], []byte(css.String()), 0666)
}

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
