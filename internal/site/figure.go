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
	Draft  bool    // draft: true; TODO comments show (see todo.go)
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
// Images in a page bundle are also pointed at their resized copies, with a
// srcset, and load lazily. An attribute block right after an image changes
// that, for a photo near the top of the page:
//
//	![A friend on the trail](trail.jpg){loading=eager}
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
	sizes   string
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

	for _, img := range images {
		attrs, err := takeImageAttrs(img, reader.Source())
		if err == nil && page != nil && page.Bundle != "" {
			err = page.usePhoto(img, attrs)
		} else if err == nil && attrs.loading != "" {
			img.SetAttributeString("loading", attrs.loading)
		}
		if err != nil {
			pc.Set(figureErrKey, err)
			return
		}
	}
	for _, p := range paragraphs {
		toFigure(p, reader.Source())
	}
}

var figureErrKey = parser.NewContextKey()

// imageAttrs are the settings from an attribute block after an image.
type imageAttrs struct {
	loading string // "eager" or "lazy", or "" when not set
}

// takeImageAttrs reads an attribute block like {loading=eager} written right
// after img, and removes it from the text so it isn't published.
func takeImageAttrs(img *ast.Image, src []byte) (imageAttrs, error) {
	var attrs imageAttrs
	t, ok := img.NextSibling().(*ast.Text)
	if !ok {
		return attrs, nil
	}
	value := string(t.Value(src))
	if !strings.HasPrefix(value, "{") {
		return attrs, nil
	}
	end := strings.Index(value, "}")
	if end < 0 {
		return attrs, fmt.Errorf("image %q: attribute block has no closing }", img.Destination)
	}
	for _, field := range strings.Fields(value[1:end]) {
		key, val, _ := strings.Cut(field, "=")
		switch {
		case key == "loading" && (val == "eager" || val == "lazy"):
			attrs.loading = val
		default:
			return attrs, fmt.Errorf("image %q: unknown attribute %q (supported: loading=eager, loading=lazy)", img.Destination, field)
		}
	}
	t.Segment = t.Segment.WithStart(t.Segment.Start + end + 1)
	return attrs, nil
}

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
func (page *pageInfo) usePhoto(img *ast.Image, attrs imageAttrs) error {
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

	// Photos load lazily unless marked {loading=eager}, for one near the top of
	// the page that shouldn't wait for layout to find it on screen.
	sizes := photoSizes
	if attrs.loading == "eager" {
		sizes = photoSizesEager
	}

	src := ph.Fallback.Largest()
	img.Destination = []byte(src.URL)
	if srcset := ph.Fallback.Srcset(); srcset != "" {
		img.SetAttributeString("srcset", srcset)
		img.SetAttributeString("sizes", sizes)
	}
	if src.Width > 0 {
		img.SetAttributeString("width", strconv.Itoa(src.Width))
		img.SetAttributeString("height", strconv.Itoa(src.Height))
	}
	if sizes == photoSizes {
		img.SetAttributeString("loading", "lazy")
	}
	img.SetAttributeString("decoding", "async")

	if len(ph.Sources) > 0 {
		pic := &pictureNode{sources: ph.Sources, sizes: sizes}
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
		pic := n.(*pictureNode)
		for _, s := range pic.sources {
			fmt.Fprintf(w, "<source type=\"%s\" srcset=\"%s\" sizes=\"%s\">",
				s.Type, html.EscapeString(s.Srcset()), pic.sizes)
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
