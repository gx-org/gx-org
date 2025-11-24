package editor

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gx-org/gx-org/internal/history"
	"github.com/gx-org/gx-org/internal/wasm/codefmt"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type state struct {
	src string
	sel *selection
}

func (s state) String() string {
	line, column := -1, -1
	if s.sel != nil {
		line, column = s.sel.Line(), s.sel.Column()
	}
	return fmt.Sprintf("%d:%d:%s", line, column, s.src)
}

func stateEq(a, b state) bool {
	return a.src == b.src
}

type Editor struct {
	gui       *ui.UI
	input     *dom.HTMLDivElement
	source    *history.History[state]
	formatter *codefmt.Formatter

	updateCode, runCode func(string)
	lastUpdate          int
}

func New(gui *ui.UI, parent dom.Element, updateCode, runCode func(src string)) *Editor {
	ed := &Editor{
		source:     history.New(stateEq),
		formatter:  codefmt.Go(),
		updateCode: updateCode,
		runCode:    runCode,
	}
	ed.input = gui.CreateDIV(
		parent,
		ui.Class("code_source_textinput"),
		ui.Property("contenteditable", "true"),
		ui.Listener("input", ed.onSourceChange),
		ui.Listener("paste", ed.onPaste),
		ui.KeyListener(ed.onKeyPress),
	)
	return ed
}

func insertSource(src string, sel *selection, toInsert string) (string, *selection, bool) {
	cursorLine := sel.Line()
	var targetLines []string
	srcLines := strings.Split(src, "\n")
	for currentSrcLine, srcLine := range srcLines {
		if currentSrcLine < cursorLine {
			targetLines = append(targetLines, srcLine)
			continue
		}
		if currentSrcLine > cursorLine {
			targetLines = append(targetLines, srcLine)
			continue
		}
		srcLineRunes := []rune(srcLine)
		cursorColumn := min(sel.Column(), len(srcLineRunes))
		newLine := slices.Clone(srcLineRunes[:cursorColumn])
		for insertedLine := range strings.Lines(toInsert) {
			newLine = append(newLine, []rune(strings.TrimSuffix(insertedLine, "\n"))...)
			if strings.HasSuffix(insertedLine, "\n") {
				targetLines = append(targetLines, string(newLine))
				newLine = []rune{}
				sel.MoveToNextLine()
			} else {
				sel.MoveColumnBy(insertedLine)
			}
		}
		newLine = append(newLine, srcLineRunes[cursorColumn:]...)
		targetLines = append(targetLines, string(newLine))
	}
	out := strings.Join(targetLines, "\n")
	return out, sel, true
}

func (ed *Editor) insertSource(inserted string) func(src string, sel *selection) (string, *selection, bool) {
	return func(src string, sel *selection) (string, *selection, bool) {
		return insertSource(src, sel, inserted)
	}
}

func (ed *Editor) onPaste(ev *dom.ClipboardEvent) {
	ev.PreventDefault()
	txt := ev.ClipboardData().GetData("text/plain")
	ed.updateSource(ed.insertSource(txt))
}

const tabSize = 4

var tabSpaces = strings.Repeat(" ", tabSize)

func (ed *Editor) onKeyPress(keys *ui.Keys, ev *dom.KeyboardEvent) {
	if keys.On("Shift") && keys.On("Enter") {
		// Run the code
		ev.PreventDefault()
		ed.runCode(ed.Text())
		return
	}
	if keys.On("Enter") {
		ev.PreventDefault()
		ed.updateSource(ed.insertSource("\n"))
		return
	}
	if keys.On("Tab") {
		ev.PreventDefault()
		ed.updateSource(ed.insertSource(tabSpaces))
		return
	}
	if (keys.On("Meta") || keys.On("Control")) && keys.On("z") {
		ed.updateSource(func(string, *selection) (string, *selection, bool) {
			if keys.On("Shift") {
				ed.source.Redo()
			} else {
				ed.source.Undo()
			}
			current := ed.source.Current()
			return current.src, current.sel, true
		})
		ev.PreventDefault()
		return
	}
}

func (ed *Editor) extractSource() string {
	src := textContent(ed.input)
	if len(src) > 0 {
		src += "\n"
	}
	return src
}

func customizeChromaHTMLSource(src string) string {
	src = strings.ReplaceAll(src,
		"<span class=\"chroma_w\">\n</span>",
		"<span class=\"chroma_w\"><br></span>",
	)
	return src
}

func (ed *Editor) Set(src string) {
	ed.set(src, nil)
}

func (ed *Editor) Text() string {
	return ed.source.Current().src
}

func (ed *Editor) set(src string, sel *selection) {
	ed.source.Append(state{src: src, sel: sel})
	parent := ed.input
	ui.ClearChildren(parent)
	formattedSrc := ed.formatter.Format(src)
	formattedSrc = customizeChromaHTMLSource(formattedSrc)
	parent.SetInnerHTML(formattedSrc)
	if sel != nil {
		sel.SetAsCurrent()
	}
	ed.updateCode(src)
}

func (ed *Editor) updateSource(process func(src string, sel *selection) (string, *selection, bool)) {
	currentSrc := ed.extractSource()
	sel := currentSelection(ed.gui, ed.input)
	currentSrc, sel, cont := process(currentSrc, sel)
	if !cont {
		return
	}
	ed.set(currentSrc, sel)
}

func (ed *Editor) onSourceChange(dom.Event) {
	ed.updateSource(func(src string, sel *selection) (string, *selection, bool) {
		return src, sel, ed.Text() != src
	})
}
