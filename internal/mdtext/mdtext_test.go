package mdtext_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/gx-org/gx-org/internal/mdtext"
)

//go:embed test01.mdl
var test01 []byte

func simplify(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimSpace(s)
	return s
}

func TestParse(t *testing.T) {
	tests := []struct {
		wantHTML  string
		wantTitle string
		md        []byte
		code      map[string]string
	}{
		{ /*Empty source*/ },
		{
			md: test01,
			code: map[string]string{
				"main":   "some code for main",
				"config": "some code for config",
			},
			wantHTML: `<h1 id="title-1">Title 1</h1>

<p>Some text</p>
`,
		},
	}
	for i, test := range tests {
		mdText := mdtext.Parse(test.md)
		if mdText.HTML != test.wantHTML {
			t.Errorf("unexpected HTML in test %d:\ngot:\n%s\nwant:\n%s\n", i, mdText.HTML, test.wantHTML)
		}
		for tag, codeWant := range test.code {
			codeGot := simplify(mdText.Code[tag])
			if codeGot != codeWant {
				t.Errorf("unexpected GX code for tag %s in test %d:\ngot:\n%s\nwant:\n%s\n", tag, i, codeGot, codeWant)
			}
		}
	}
}
