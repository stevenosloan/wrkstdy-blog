package site

import (
	"fmt"
	"html"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// pageInfo is what the markdown renderer needs to know about the page it's
// rendering, passed through goldmark's parser context.
type pageInfo struct {
	URL    string  // where the page is published, like "/journal/lake-day/"
	Bundle string  // the page's folder when it's a bundle (folder/index.md), else ""
	photos *photos // resizes the bundle's images
}

var pageInfoKey = parser.NewContextKey()

// figureExtension turns a paragraph that starts with an image into a <figure>.
// Text on the lines right under the image (no blank line between) becomes its
// <figcaption>:
//
//	![Rowboat tied to the dock](dock.jpg)
//	First morning: fog didn't lift until ten.
//
// Images in a page bundle are also pointed at their resized copies, with a srcset.
type figureExtension struct{}

func (figureExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(figureTransformer{}, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(figureRenderer{}, 100)))
}

var (
	kindFigure     = ast.NewNodeKind("Figure")
	kindFigcaption = ast.NewNodeKind("Figcaption")
	kindPicture    = ast.NewNodeKind("Picture")
)

// pictureNode wraps a resized photo's <img> in a <picture>, with a <source> for
// each of its other formats.
type pictureNode struct {
	ast.BaseInline
	sources []photoSet
}

func (n *pictureNode) Kind() ast.NodeKind { return kindPicture }
func (n *pictureNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

type figureNode struct{ ast.BaseBlock }

func (n *figureNode) Kind() ast.NodeKind { return kindFigure }
func (n *figureNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

type figcaptionNode struct{ ast.BaseBlock }

func (n *figcaptionNode) Kind() ast.NodeKind { return kindFigcaption }
func (n *figcaptionNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, nil, nil)
}

type figureTransformer struct{}

func (figureTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	page, _ := pc.Get(pageInfoKey).(*pageInfo)

	var paragraphs []*ast.Paragraph
	var images []*ast.Image
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Paragraph:
			if _, ok := n.FirstChild().(*ast.Image); ok {
				paragraphs = append(paragraphs, n)
			}
		case *ast.Image:
			images = append(images, n)
		}
		return ast.WalkContinue, nil
	})

	if page != nil && page.Bundle != "" {
		for _, img := range images {
			if err := page.usePhoto(img); err != nil {
				pc.Set(figureErrKey, err)
				return
			}
		}
	}
	for _, p := range paragraphs {
		toFigure(p, reader.Source())
	}
}

var figureErrKey = parser.NewContextKey()

// toFigure replaces p with a figure holding its image, and its remaining lines
// as a caption when the image is followed by a line break. An image with more
// text on the same line is left as an inline image.
func toFigure(p *ast.Paragraph, src []byte) {
	img := p.FirstChild()
	next := img.NextSibling()
	if next != nil {
		t, ok := next.(*ast.Text)
		if !ok || !t.SoftLineBreak() || len(strings.TrimSpace(string(t.Value(src)))) > 0 {
			return
		}
	}

	fig := &figureNode{}
	p.Parent().ReplaceChild(p.Parent(), p, fig)
	fig.AppendChild(fig, img)
	if next == nil {
		return
	}
	caption := &figcaptionNode{}
	for n := next.NextSibling(); n != nil; {
		after := n.NextSibling()
		caption.AppendChild(caption, n)
		n = after
	}
	if caption.HasChildren() {
		fig.AppendChild(fig, caption)
	}
}

// usePhoto points img at its file in the page bundle, resized, when its
// destination is a relative path. Links elsewhere are left alone.
func (page *pageInfo) usePhoto(img *ast.Image) error {
	dest := string(img.Destination)
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || strings.HasPrefix(dest, "/") || dest == "" {
		return nil
	}
	rel := filepath.Clean(filepath.FromSlash(u.Path))
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("image %q is outside the page's folder", dest)
	}
	ph, err := page.photos.add(filepath.Join(page.Bundle, rel), page.URL)
	if err != nil {
		return fmt.Errorf("image %q: %w", dest, err)
	}

	src := ph.Fallback.Largest()
	img.Destination = []byte(src.URL)
	if srcset := ph.Fallback.Srcset(); srcset != "" {
		img.SetAttributeString("srcset", srcset)
		img.SetAttributeString("sizes", photoSizes)
	}
	if src.Width > 0 {
		img.SetAttributeString("width", strconv.Itoa(src.Width))
		img.SetAttributeString("height", strconv.Itoa(src.Height))
	}
	img.SetAttributeString("loading", "lazy")
	img.SetAttributeString("decoding", "async")

	if len(ph.Sources) > 0 {
		pic := &pictureNode{sources: ph.Sources}
		img.Parent().ReplaceChild(img.Parent(), img, pic)
		pic.AppendChild(pic, img)
	}
	return nil
}

type figureRenderer struct{}

func (r figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindFigure, func(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			w.WriteString("<figure>\n")
		} else {
			w.WriteString("\n</figure>\n")
		}
		return ast.WalkContinue, nil
	})
	reg.Register(kindPicture, func(w util.BufWriter, _ []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			w.WriteString("</picture>")
			return ast.WalkContinue, nil
		}
		w.WriteString("<picture>")
		for _, s := range n.(*pictureNode).sources {
			fmt.Fprintf(w, "<source type=\"%s\" srcset=\"%s\" sizes=\"%s\">",
				s.Type, html.EscapeString(s.Srcset()), photoSizes)
		}
		return ast.WalkContinue, nil
	})
	reg.Register(kindFigcaption, func(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			w.WriteString("\n<figcaption>")
		} else {
			w.WriteString("</figcaption>")
		}
		return ast.WalkContinue, nil
	})
}
