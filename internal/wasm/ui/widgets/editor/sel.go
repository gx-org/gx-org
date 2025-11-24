package editor

import (
	"fmt"
	"math"
	"strings"
	"syscall/js"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type selection struct {
	ui          *ui.UI
	el          dom.HTMLElement
	line        int
	utf16Column int
	utf8Column  int
	rang        js.Value
}

func jsSelection() js.Value {
	return js.Global().Call("getSelection")
}

func isLineNode(node dom.Node) bool {
	for _, class := range ui.ClassOf(node) {
		if strings.HasSuffix(class, "line") {
			return true
		}
	}
	return false
}

func findParentLineAndCode(el dom.Element) (line, code dom.Element) {
	current := el
	for current != nil && ui.NodeName(current) != "CODE" {
		if isLineNode(current) {
			line = current
		}
		current = current.ParentElement()
	}
	code = current
	return
}

func lineNumFromElement(el dom.Element) (dom.Element, int) {
	lineEl, codeEl := findParentLineAndCode(el)
	if codeEl == nil || lineEl == nil {
		return nil, 0
	}
	line := 0
	prev := lineEl.PreviousElementSibling()
	for prev != nil {
		line++
		prev = prev.PreviousElementSibling()
	}
	return lineEl, line
}

func utf16Count(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func textLenFromPreviousElement(ancestor dom.Element, line dom.Element) (utf16Pos, utf8Pos int) {
	if line == nil {
		line = ancestor
	}
	const id = "ancestor_mark"
	ancestor.SetID(id)
	text := textContentUntil(line, func(node dom.Node) bool {
		return dom.WrapElement(node.Underlying()).ID() != id
	})
	return utf16Count(text), utf8.RuneCountInString(text)
}

func textLenFromElement(rang js.Value, ancestor dom.Element) (utf16Pos, utf8Pos int) {
	utf16Pos = rang.Get("startOffset").Int()
	utf8Str := textContentUntil(ancestor, noFilter)
	utf16Str := utf16.Encode([]rune(utf8Str))
	if utf16Pos > len(utf16Str) {
		return 0, 0
	}
	subUTF16 := utf16Str[:utf16Pos]
	utf8Pos = len(utf16.Decode(subUTF16))
	return
}

func currentSelection(gui *ui.UI, el dom.HTMLElement) *selection {
	sel := &selection{
		ui: gui,
		el: el,
	}
	if numRange := jsSelection().Get("rangeCount").Int(); numRange == 0 {
		return sel
	}
	sel.rang = jsSelection().Call("getRangeAt", 0)
	ancestor := dom.WrapElement(sel.rang.Get("commonAncestorContainer"))
	if ancestor == nil {
		return sel
	}
	var lineEl dom.Element
	lineEl, sel.line = lineNumFromElement(ancestor)
	utf16Prev, utf8Prev := textLenFromPreviousElement(ancestor, lineEl)
	utf16Column, utf8Column := textLenFromElement(sel.rang, ancestor)
	sel.utf16Column = utf16Prev + utf16Column
	sel.utf8Column = utf8Prev + utf8Column
	return sel
}

func findChild(el dom.Node, filter func(dom.Node) bool) dom.Node {
	if filter(el) {
		return el
	}
	for _, child := range el.ChildNodes() {
		if found := findChild(child, filter); found != nil {
			return found
		}
	}
	return nil
}

func computeChildColumn(line dom.Node, until int) (dom.Node, int) {
	column := until
	last := line
	var lastLen int
	for child := range ui.IterLeaves(line, noFilter) {
		textLen := utf16Count(textContentUntil(child, noFilter))
		if column <= textLen {
			return child, column
		}
		column -= textLen
		if column < 0 {
			break
		}
		last = child
		lastLen = textLen
	}
	return last, lastLen
}

func findFirstLeaf(el dom.Node) dom.Node {
	for leaf := range ui.IterLeaves(el, noFilter) {
		return leaf
	}
	return nil
}

func (sel *selection) SetAsCurrent() {
	if sel.rang.IsNull() {
		return
	}
	codeNode := findChild(sel.el, func(node dom.Node) bool {
		return ui.NodeName(node) == "CODE"
	})
	if codeNode == nil {
		return
	}
	lineEls := codeNode.ChildNodes()
	if len(lineEls) == 0 {
		return
	}
	if sel.line >= len(lineEls) {
		sel.line = len(lineEls) - 1
		sel.utf16Column = math.MaxInt
	}
	child, column := computeChildColumn(lineEls[sel.line], sel.utf16Column)
	jsSelection().Call("collapse", findFirstLeaf(child).Underlying(), column)
}

func (sel *selection) Line() int {
	return sel.line
}

func (sel *selection) Column() int {
	return sel.utf8Column
}

func (sel *selection) MoveColumnBy(s string) {
	sel.utf16Column += utf16Count(s)
	sel.utf8Column += utf8.RuneCountInString(s)
}

func (sel *selection) MoveToNextLine() {
	sel.utf16Column = 0
	sel.utf8Column = 0
	sel.line++
}

func (sel *selection) String() string {
	if sel == nil {
		return "nil"
	}
	return fmt.Sprintf("line: %d col: %d", sel.line, sel.utf16Column)
}
