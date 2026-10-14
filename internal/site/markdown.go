package site

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.Typographer, alertExtension{}),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			// Lower numbers run first, so this replaces goldmark's own code block renderer (1000).
			renderer.WithNodeRenderers(util.Prioritized(codeBlockRenderer{}, 100)),
		),
	)
}

// langAliases maps fence names to MicroLighter grammar names it doesn't alias itself.
var langAliases = map[string]string{
	"diff":    "git-diff",
	"patch":   "git-diff",
	"golang":  "go",
	"console": "bash",
}

// codeBlockRenderer writes fenced code as plain <pre><code class="language-x"> for
// MicroLighter to color in the browser, and ```mermaid blocks as <pre class="mermaid">.
type codeBlockRenderer struct{}

func (r codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.render)
}

func (codeBlockRenderer) render(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	block := n.(*ast.FencedCodeBlock)

	var code strings.Builder
	lines := block.Lines()
	for i := range lines.Len() {
		seg := lines.At(i)
		code.Write(seg.Value(src))
	}
	body := html.EscapeString(code.String())

	lang := string(block.Language(src))
	if lang == "mermaid" {
		fmt.Fprintf(w, "<pre class=\"mermaid\">%s</pre>\n", body)
		return ast.WalkSkipChildren, nil
	}
	if alias, ok := langAliases[lang]; ok {
		lang = alias
	}
	if lang == "" {
		fmt.Fprintf(w, "<pre><code>%s</code></pre>\n", body)
	} else {
		fmt.Fprintf(w, "<pre><code class=\"language-%s\">%s</code></pre>\n", html.EscapeString(lang), body)
	}
	return ast.WalkSkipChildren, nil
}

// shortcodeRE matches {{< name >}}, the same syntax Hugo uses.
var shortcodeRE = regexp.MustCompile(`\{\{<\s*([\w-]+)\s*>\}\}`)

// renderMarkdown turns a page body into HTML. Shortcodes are swapped for
// placeholder words before goldmark runs (it would escape their HTML), and the
// rendered templates/shortcodes/<name>.html goes back in afterwards.
func (b *builder) renderMarkdown(body []byte) (string, error) {
	var snippets []string
	var scErr error
	body = shortcodeRE.ReplaceAllFunc(body, func(m []byte) []byte {
		name := string(shortcodeRE.FindSubmatch(m)[1])
		out, err := b.renderShortcode(name)
		if err != nil {
			scErr = err
		}
		snippets = append(snippets, string(out))
		return fmt.Appendf(nil, "WRKSTDYSHORTCODE%dEND", len(snippets)-1)
	})
	if scErr != nil {
		return "", scErr
	}

	var buf bytes.Buffer
	if err := b.md.Convert(body, &buf); err != nil {
		return "", err
	}
	out := buf.String()
	for i, s := range snippets {
		token := fmt.Sprintf("WRKSTDYSHORTCODE%dEND", i)
		out = strings.Replace(out, "<p>"+token+"</p>", s, 1)
		out = strings.Replace(out, token, s, 1)
	}
	return out, nil
}

func (b *builder) renderShortcode(name string) (template.HTML, error) {
	t := b.shortcodes.Lookup(name + ".html")
	if t == nil {
		return "", fmt.Errorf("unknown shortcode %q", name)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, nil); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
