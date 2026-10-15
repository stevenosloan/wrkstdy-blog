package site

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFigures(t *testing.T) {
	b := &builder{md: newMarkdown()}
	tests := []struct{ name, src, want string }{
		{"caption from the lines under it", "![Dock](/dock.jpg)\nFirst *morning*.",
			"<figure>\n<img src=\"/dock.jpg\" alt=\"Dock\">\n<figcaption>First <em>morning</em>.</figcaption>\n</figure>\n"},
		{"image alone", "![Dock](/dock.jpg)",
			"<figure>\n<img src=\"/dock.jpg\" alt=\"Dock\">\n</figure>\n"},
		{"blank line means no caption", "![Dock](/dock.jpg)\n\nA paragraph.",
			"<figure>\n<img src=\"/dock.jpg\" alt=\"Dock\">\n</figure>\n<p>A paragraph.</p>\n"},
		{"text on the same line stays inline", "![Dock](/dock.jpg) and more",
			"<p><img src=\"/dock.jpg\" alt=\"Dock\"> and more</p>\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.renderMarkdown([]byte(tt.src), nil)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

// writeJPEG writes a w×h test photo to path.
func writeJPEG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		img.Set(x, x%h, color.RGBA{R: 200, A: 255})
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
}

func TestBundlePhotos(t *testing.T) {
	root, out := t.TempDir(), t.TempDir()
	dir := filepath.Join(root, "content", "journal", "2026-10-15-lake-day")
	writeJPEG(t, filepath.Join(dir, "dock.jpg"), 1000, 500)
	writeJPEG(t, filepath.Join(dir, "unused.jpg"), 100, 100)
	md := "---\ntitle: Lake day\n---\n\n![Dock](dock.jpg)\nFog until ten.\n"
	if err := writeFile(filepath.Join(dir, "index.md"), []byte(md)); err != nil {
		t.Fatal(err)
	}

	writeFile(filepath.Join(root, "content", "_index.md"), []byte("---\ntitle: home\n---\n"))
	writeFile(filepath.Join(root, "content", "journal", "_index.md"), []byte("---\ntitle: journal\n---\n"))

	b := &builder{md: newMarkdown(), photos: newPhotos(filepath.Join(root, ".cache"))}
	c, err := b.loadContent(filepath.Join(root, "content"))
	if err != nil {
		t.Fatal(err)
	}
	page := c.sections["journal"].Pages[0]
	if page.URL != "/journal/lake-day/" || !page.Date.Equal(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("URL %q date %v, want the folder's name to give both", page.URL, page.Date)
	}
	html := string(page.Content)
	for _, want := range []string{
		"<figure>\n" + `<picture><source type="image/avif" srcset="/journal/lake-day/dock.640w.avif 640w, /journal/lake-day/dock.960w.avif 960w" sizes="` + photoSizes + `">`,
		`<source type="image/webp" srcset="/journal/lake-day/dock.640w.webp 640w, /journal/lake-day/dock.960w.webp 960w"`,
		`src="/journal/lake-day/dock.960w.jpg"`,
		`srcset="/journal/lake-day/dock.640w.jpg 640w, /journal/lake-day/dock.960w.jpg 960w"`,
		`decoding="async"></picture>`,
		`sizes="` + photoSizes + `"`,
		`width="960" height="480"`,
		`loading="lazy"`,
		"<figcaption>Fog until ten.</figcaption>",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("page HTML is missing %s:\n%s", want, html)
		}
	}

	if err := b.photos.writeTo(out); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(out, "journal", "lake-day", "*"))
	if len(files) != 6 {
		t.Errorf("published %v, want only two sizes of dock.jpg in three formats (not the original, nor the unused photo)", files)
	}
	for _, ext := range []string{".jpg", ".webp", ".avif"} {
		cfg, err := decodeConfig(filepath.Join(out, "journal", "lake-day", "dock.640w"+ext))
		if err != nil || cfg.Width != 640 || cfg.Height != 320 {
			t.Errorf("dock.640w%s is %dx%d (%v), want 640x320", ext, cfg.Width, cfg.Height, err)
		}
	}

	// A missing photo fails the build rather than publishing a broken image.
	writeFile(filepath.Join(dir, "index.md"), []byte("![Gone](gone.jpg)\n"))
	if _, err := b.loadContent(filepath.Join(root, "content")); err == nil || !strings.Contains(err.Error(), "gone.jpg") {
		t.Errorf("err = %v, want one naming gone.jpg", err)
	}
	writeFile(filepath.Join(dir, "index.md"), []byte("![Up](../secret.jpg)\n"))
	if _, err := b.loadContent(filepath.Join(root, "content")); err == nil {
		t.Error("want an error for a photo outside the bundle")
	}
}

func TestVariantWidths(t *testing.T) {
	if got := variantWidths(4000); len(got) != len(photoWidths) {
		t.Errorf("variantWidths(4000) = %v, want every width", got)
	}
	if got := variantWidths(1000); len(got) != 2 || got[1] != 960 {
		t.Errorf("variantWidths(1000) = %v, want [640 960]", got)
	}
	if got := variantWidths(300); len(got) != 1 || got[0] != 300 {
		t.Errorf("variantWidths(300) = %v, want [300], the photo's own width", got)
	}
}

func TestAbsLinksSrcset(t *testing.T) {
	got := absLinks(`<img src="/a.960w.jpg" srcset="/a.640w.jpg 640w, /a.960w.jpg 960w">`, "https://x.com")
	want := `<img src="https://x.com/a.960w.jpg" srcset="https://x.com/a.640w.jpg 640w, https://x.com/a.960w.jpg 960w">`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
