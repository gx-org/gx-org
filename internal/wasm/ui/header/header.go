package header

import (
	"github.com/gx-org/gx-org/internal/lessons"
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Header struct {
	el dom.HTMLElement
}

func New(gui *ui.UI) *Header {
	els := gui.Dom().Document().GetElementsByTagName("header")
	hdr := &Header{}
	if len(els) == 0 {
		return hdr
	}
	hdr.el = els[0].(dom.HTMLElement)
	hdr.SetContent(nil)
	return hdr
}

func (hdr *Header) SetContent(les *lessons.Lesson) {
	pageTitle := "Walkthrough"
	if les != nil && les.Options.ChapterTitleAsPageTitle {
		pageTitle = les.Options.ChapterTitle
	}
	hdr.el.SetInnerHTML(`<a href="index.html"><img src="res/gxlogo.png" style="margin-left: 15px; margin-right: 15px; float: left; height:100%; object-fit: contain;" alt="GX Logo"></a>` + pageTitle)
}
