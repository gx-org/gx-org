package mdtext_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gx-org/gx-org/internal/mdtext"
)

func TestParse(t *testing.T) {
	tests := []struct {
		wantHTML  string
		wantTitle string
		md        string
		code      map[string]string
	}{
		{ /*Empty source*/ },
		{
			md: `
# Title 1

Some text
`,
			code: map[string]string{
				mdtext.TagPrefix + "code": `
some code
`,
			},
			wantTitle: `<h1 id="title-1">Title 1</h1>
`,
			wantHTML: `<p>Some text</p>
`,
		},
	}
	for i, test := range tests {
		var mdSrc strings.Builder
		mdSrc.WriteString(test.md)
		mdSrc.WriteString("\n")
		for tag, code := range test.code {
			mdSrc.WriteString(fmt.Sprintf("```%s\n%s```\n", tag, code))
		}
		mdText := mdtext.Parse([]byte(mdSrc.String()))
		if mdText.TitleHTML != test.wantTitle {
			t.Errorf("unexpected title in test %d:\ngot:\n%s\nwant:\n%s\n", i, mdText.TitleHTML, test.wantTitle)
		}
		if mdText.HTML != test.wantHTML {
			t.Errorf("unexpected HTML in test %d:\ngot:\n%s\nwant:\n%s\n", i, mdText.HTML, test.wantHTML)
		}
		for tag, codeWant := range test.code {
			codeGot := mdText.Code[tag]
			if codeGot != codeWant {
				t.Errorf("unexpected GX code for tag %s in test %d:\ngot:\n%s\nwant:\n%s\n", tag, i, codeGot, codeWant)
			}
		}
	}
}
