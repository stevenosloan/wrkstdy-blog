package site

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// TODO comments: on a draft page, an HTML comment on its own, like
//
//	<!-- TODO(steven): the opening. -->
//
// renders as an alert with a "TODO" title, so the gaps show while previewing.
// On a published page it stays a comment, which goldmark leaves out.

// todoRE matches the whole comment; the (name) and colon are optional.
var todoRE = regexp.MustCompile(`(?s)^<!--\s*TODO(?:\([^)]*\))?:?\s*(.*?)\s*-->\s*$`)

type todoExtension struct{}

func (todoExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(todoTransformer{}, 100)))
}

type todoTransformer struct{}

func (todoTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	page, _ := pc.Get(pageInfoKey).(*pageInfo)
	if page == nil || !page.Draft {
		return
	}
	src := reader.Source()
	var blocks []*ast.HTMLBlock
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if b, ok := n.(*ast.HTMLBlock); ok && entering && b.HTMLBlockType == ast.HTMLBlockType2 {
			blocks = append(blocks, b)
		}
		return ast.WalkContinue, nil
	})
	// Swap after walking, as in alerts.go.
	for _, b := range blocks {
		var raw bytes.Buffer
		for i := range b.Lines().Len() {
			seg := b.Lines().At(i)
			raw.Write(seg.Value(src))
		}
		if b.HasClosure() {
			raw.Write(b.ClosureLine.Value(src))
		}
		m := todoRE.FindSubmatch(raw.Bytes())
		if m == nil {
			continue
		}
		alert := &alertNode{kind: "todo"}
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString(m[1]))
		alert.AppendChild(alert, para)
		b.Parent().ReplaceChild(b.Parent(), b, alert)
	}
}
