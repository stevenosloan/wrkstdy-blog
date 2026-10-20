package site

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gen2brain/webp"
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
		{"attribute block is applied and removed", "![Dock](/dock.jpg){loading=eager}\nFirst *morning*.",
			"<figure>\n<img src=\"/dock.jpg\" alt=\"Dock\" loading=\"eager\">\n<figcaption>First <em>morning</em>.</figcaption>\n</figure>\n"},
		{"attribute block before more text", "![Dock](/dock.jpg){loading=lazy} and more",
			"<p><img src=\"/dock.jpg\" alt=\"Dock\" loading=\"lazy\"> and more</p>\n"},
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

func TestImageAttrErrors(t *testing.T) {
	b := &builder{md: newMarkdown()}
	for _, src := range []string{"![Dock](/dock.jpg){loading=eagre}", "![Dock](/dock.jpg){width=10}", "![Dock](/dock.jpg){loading=eager"} {
		if _, err := b.renderMarkdown([]byte(src), nil); err == nil {
			t.Errorf("render(%q): want an error", src)
		}
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
	if err := writeFile(filepath.Join(dir, "icon.svg"), []byte("<svg/>")); err != nil {
		t.Fatal(err)
	}
	var crop bytes.Buffer
	if err := webp.Encode(&crop, image.NewRGBA(image.Rect(0, 0, 30, 20)), webp.Options{Lossless: true}); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(dir, "crop.webp"), crop.Bytes()); err != nil {
		t.Fatal(err)
	}
	md := "---\ntitle: Lake day\n---\n\n![Dock](dock.jpg){loading=eager}\nFog until ten.\n\n![Dock again](dock.jpg)\n\n![Icon](icon.svg)\n\n![Crop](crop.webp)\n"
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
	// Every published photo has a content hash in its name. Checked here, then
	// taken out so the expected HTML below can be written plainly.
	html := string(page.Content)
	if !regexp.MustCompile(`"/assets/photos/journal/lake-day/dock\.960w\.[0-9a-f]{12}\.jpg"`).MatchString(html) {
		t.Errorf("want photos under /assets/photos with a hash in their names:\n%s", html)
	}
	html = photoHash.ReplaceAllString(html, ".")
	for _, want := range []string{
		"<figure>\n" + `<picture><source type="image/avif" srcset="/assets/photos/journal/lake-day/dock.640w.avif 640w, /assets/photos/journal/lake-day/dock.960w.avif 960w" sizes="` + photoSizesEager + `">`,
		`<source type="image/webp" srcset="/assets/photos/journal/lake-day/dock.640w.webp 640w, /assets/photos/journal/lake-day/dock.960w.webp 960w"`,
		`src="/assets/photos/journal/lake-day/dock.960w.jpg"`,
		`srcset="/assets/photos/journal/lake-day/dock.640w.jpg 640w, /assets/photos/journal/lake-day/dock.960w.jpg 960w"`,
		`decoding="async"></picture>`,
		`width="960" height="480"`,
		// A photo marked {loading=eager} loads straight away; the rest are lazy, with sizes="auto".
		`alt="Dock" srcset="/assets/photos/journal/lake-day/dock.640w.jpg 640w, /assets/photos/journal/lake-day/dock.960w.jpg 960w" sizes="` + photoSizesEager + `" width="960" height="480" decoding="async">`,
		`alt="Dock again" srcset="/assets/photos/journal/lake-day/dock.640w.jpg 640w, /assets/photos/journal/lake-day/dock.960w.jpg 960w" sizes="` + photoSizes + `" width="960" height="480" loading="lazy" decoding="async">`,
		"<figcaption>Fog until ten.</figcaption>",
		// Formats that aren't resized are published as they are, without a srcset.
		`<img src="/assets/photos/journal/lake-day/icon.svg" alt="Icon" loading="lazy" decoding="async">`,
		// ...with their size when Go can read it.
		`<img src="/assets/photos/journal/lake-day/crop.webp" alt="Crop" width="30" height="20" loading="lazy" decoding="async">`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("page HTML is missing %s:\n%s", want, html)
		}
	}

	if err := b.photos.writeTo(out); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(out, "assets", "photos", "journal", "lake-day", "*"))
	if len(files) != 8 {
		t.Errorf("published %v, want two sizes of dock.jpg in three formats, icon.svg and crop.webp (not the original, nor the unused photo)", files)
	}
	for _, ext := range []string{".jpg", ".webp", ".avif"} {
		matches, _ := filepath.Glob(filepath.Join(out, "assets", "photos", "journal", "lake-day", "dock.640w.*"+ext))
		if len(matches) != 1 {
			t.Fatalf("dock.640w.*%s: found %v", ext, matches)
		}
		cfg, err := decodeConfig(matches[0])
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

// photoHash matches the content hash in a published photo's name.
var photoHash = regexp.MustCompile(`\.[0-9a-f]{12}\.`)

func TestPublishedPhotoNames(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	writeFile(a, []byte("one"))
	writeFile(b, []byte("two"))
	p := newPhotos(dir)
	urlA, _ := p.publish(a, "/recipes/wings/", "seasoned.640w.avif")
	urlA2, _ := p.publish(a, "/recipes/wings/", "seasoned.640w.avif")
	urlB, _ := p.publish(b, "/recipes/wings/", "seasoned.640w.avif")
	if !regexp.MustCompile(`^/assets/photos/recipes/wings/seasoned\.640w\.[0-9a-f]{12}\.avif$`).MatchString(urlA) {
		t.Errorf("url = %q", urlA)
	}
	if urlA != urlA2 || urlA == urlB {
		t.Errorf("same content should give the same name and different content a new one: %q %q %q", urlA, urlA2, urlB)
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

func TestPrunePhotoCache(t *testing.T) {
	cache := t.TempDir()
	now := time.Now()
	old := now.Add(-cacheMaxAge - time.Hour)
	backdate := func(dir string, mtime time.Time) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(dir, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}
	backdate(filepath.Join(cache, "fresh"), now.Add(-time.Hour))
	backdate(filepath.Join(cache, "stale"), old)
	backdate(filepath.Join(cache, "tmp-new"), now)
	backdate(filepath.Join(cache, "tmp-old"), now.Add(-2*time.Hour))

	// An old entry that a build uses again is marked as used, and kept.
	p := newPhotos(cache)
	src := filepath.Join(t.TempDir(), "a.jpg")
	writeJPEG(t, src, 100, 100)
	used, err := p.cached(src)
	if err != nil {
		t.Fatal(err)
	}
	backdate(used, old)
	if _, err := p.cached(src); err != nil {
		t.Fatal(err)
	}

	if err := p.prune(now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{"fresh": true, "stale": false, "tmp-new": true, "tmp-old": false, filepath.Base(used): true} {
		_, err := os.Stat(filepath.Join(cache, name))
		if got := err == nil; got != want {
			t.Errorf("%s kept = %v, want %v", name, got, want)
		}
	}
	if err := newPhotos(filepath.Join(cache, "missing")).prune(now); err != nil {
		t.Errorf("pruning a cache that doesn't exist yet: %v", err)
	}
}
