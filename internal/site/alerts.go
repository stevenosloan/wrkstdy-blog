package site

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// GitHub-style alerts: a blockquote whose first line is only a marker, like
//
//	> [!NOTE]
//	> Worth knowing.
//
// renders as <div class="alert alert-note"> with a "Note" title, not a blockquote.

// alertRE matches the marker line. GitHub only recognizes these five types.
var alertRE = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$`)

var kindAlert = ast.NewNodeKind("Alert")

// alertNode takes the place of the blockquote and holds its children.
type alertNode struct {
	ast.BaseBlock
	kind string // lowercase: note, tip, important, warning, caution
}

func (n *alertNode) Kind() ast.NodeKind { return kindAlert }

func (n *alertNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"kind": n.kind}, nil)
}

type alertExtension struct{}

func (alertExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(alertTransformer{}, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(alertRenderer{}, 100)))
}

type alertTransformer struct{}

func (alertTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	var quotes []*ast.Blockquote
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if q, ok := n.(*ast.Blockquote); ok && entering {
			quotes = append(quotes, q)
		}
		return ast.WalkContinue, nil
	})
	// Swap after walking, since replacing nodes mid-walk would derail it.
	for _, q := range quotes {
		para, ok := q.FirstChild().(*ast.Paragraph)
		if !ok || para.Lines().Len() == 0 {
			continue
		}
		first := para.Lines().At(0)
		m := alertRE.FindSubmatch(first.Value(src))
		if m == nil {
			continue
		}

		// Drop the marker's inline nodes (everything that starts on the first line).
		for c := para.FirstChild(); c != nil; {
			next := c.NextSibling()
			t, ok := c.(*ast.Text)
			if !ok || t.Segment.Start >= first.Stop {
				break
			}
			para.RemoveChild(para, c)
			c = next
		}
		if para.ChildCount() == 0 {
			q.RemoveChild(q, para)
		}

		alert := &alertNode{kind: strings.ToLower(string(m[1]))}
		for c := q.FirstChild(); c != nil; {
			next := c.NextSibling()
			alert.AppendChild(alert, c)
			c = next
		}
		q.Parent().ReplaceChild(q.Parent(), q, alert)
	}
}

type alertRenderer struct{}

func (r alertRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindAlert, r.render)
}

func (alertRenderer) render(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	a := n.(*alertNode)
	if entering {
		title := strings.ToUpper(a.kind[:1]) + a.kind[1:]
		fmt.Fprintf(w, "<div class=\"alert alert-%s\">\n<p class=\"alert-title\">%s</p>\n", a.kind, title)
	} else {
		w.WriteString("</div>\n")
	}
	return ast.WalkContinue, nil
}
