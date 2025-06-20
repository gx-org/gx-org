package mdtext

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const TagPrefix = "overview:"

func processCodeWithGXTags(m map[string]*ast.CodeBlock) func(node *ast.CodeBlock) ast.WalkStatus {
	return func(node *ast.CodeBlock) ast.WalkStatus {
		codeTag := string(node.Info)
		if !strings.HasPrefix(codeTag, TagPrefix) {
			return ast.GoToNext
		}
		m[codeTag] = node
		return ast.GoToNext
	}
}

func titleNode(heading **ast.Heading) func(node *ast.Heading) ast.WalkStatus {
	return func(node *ast.Heading) ast.WalkStatus {
		if node.Level != 1 {
			return ast.GoToNext
		}
		*heading = node
		return ast.GoToNext
	}
}

func walk[T ast.Node](process func(T) ast.WalkStatus) ast.NodeVisitorFunc {
	return func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}
		nodeT, ok := node.(T)
		if !ok {
			return ast.GoToNext
		}
		return process(nodeT)
	}
}

type MDText struct {
	TitleHTML string
	Code      map[string]string
	HTML      string
}

func Parse(src []byte) *MDText {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(src)
	codeBlockTags := make(map[string]*ast.CodeBlock)
	ast.Walk(doc, walk(processCodeWithGXTags(codeBlockTags)))
	mdt := &MDText{Code: make(map[string]string)}
	for tag, codeBlock := range codeBlockTags {
		mdt.Code[tag] = string(codeBlock.Literal)
		ast.RemoveFromTree(codeBlock)
	}
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	var title *ast.Heading
	ast.Walk(doc, walk(titleNode(&title)))
	if title != nil {
		mdt.TitleHTML = string(markdown.Render(title, renderer))
		ast.RemoveFromTree(title)
	}
	mdt.HTML = string(markdown.Render(doc, renderer))
	return mdt
}
