package header

import (
	"github.com/gx-org/gx-org/internal/wasm/ui"
	"honnef.co/go/js/dom/v2"
)

type Header struct {
	el dom.HTMLElement
}

func New(gui *ui.UI) *Header {
	els := gui.Dom().Document().GetElementsByTagName("header")
	header := &Header{}
	if len(els) == 0 {
		return header
	}
	header.el = els[0].(dom.HTMLElement)
	header.el.SetInnerHTML(`<img src="res/gxlogo.png" style="margin-left: 15px; margin-right: 15px; float: left; height:100%; object-fit: contain;" alt="GX Logo">Walkthrough`)
	return header
}
