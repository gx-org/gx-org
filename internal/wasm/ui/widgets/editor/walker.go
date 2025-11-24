package editor

import (
	"html"
	"strings"

	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

func noFilter(dom.Node) bool {
	return true
}

func textContent(el dom.Node) string {
	return textContentUntil(el, noFilter)
}

type lineCollector struct {
	nextLine int
	lines    []string
	buf      strings.Builder
	all      string
}

func newLineCollector() *lineCollector {
	return &lineCollector{nextLine: -1}
}

func (lc *lineCollector) flush() {
	lc.nextLine++
	if lc.nextLine == 0 {
		return
	}
	line := lc.buf.String()
	line = strings.TrimSuffix(line, "\n")
	lc.lines = append(lc.lines, line)
	lc.buf = strings.Builder{}
}

func (lc *lineCollector) String() string {
	if lc.all != "" {
		return lc.all
	}
	if lc.buf.Len() > 0 {
		lc.nextLine = 0
	}
	lc.flush()
	lc.all = strings.Join(lc.lines, "\n")
	lc.lines = nil
	lc.nextLine = -1
	lc.buf = strings.Builder{}
	return lc.all
}

func textContentUntil(el dom.Node, filter func(dom.Node) bool) string {
	lc := newLineCollector()
	processLine := func(node dom.Node) bool {
		if isLineNode(node) {
			lc.flush()
		}
		return filter(node)
	}
	for _, data := range ui.IterLeaves(el, processLine) {
		lc.buf.WriteString(data)
	}
	return html.UnescapeString(lc.String())
}
