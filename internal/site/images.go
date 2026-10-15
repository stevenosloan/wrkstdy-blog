package site

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg" // registers JPEG with image.DecodeConfig
	_ "image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif" // also registers AVIF and WebP with image.DecodeConfig
	"github.com/gen2brain/webp"
)

// photoWidths are the widths a photo is resized to. The content column is 40rem
// (640px), so these cover screens from 1x to 3x. Widths wider than the photo are
// skipped, so nothing is upscaled.
var photoWidths = []int{640, 960, 1280, 1920}

const (
	jpegQuality = 80
	webpQuality = 70
	avifQuality = 50
	avifSpeed   = 6 // 0 (slowest, smallest) to 10; encoding AVIF is most of the time spent resizing
)

// photoFormats are the formats each width is also encoded in, best first, as
// <source>s in a <picture>. The browser takes the first it supports, and falls
// back to the <img>, a JPEG (or a PNG, for PNGs, to keep transparency).
var photoFormats = []struct {
	ext, mime string
	encode    func(io.Writer, image.Image) error
}{
	{".avif", "image/avif", func(w io.Writer, m image.Image) error {
		return avif.Encode(w, m, avif.Options{Quality: avifQuality, QualityAlpha: avifQuality, Speed: avifSpeed})
	}},
	{".webp", "image/webp", func(w io.Writer, m image.Image) error {
		return webp.Encode(w, m, webp.Options{Quality: webpQuality, Method: 4})
	}},
}

// photoSizes is the sizes="" attribute for a resized photo. "auto" lets the
// browser pick from the width the image is laid out at (it needs loading="lazy");
// browsers that don't know it yet fall back to the rest.
const photoSizes = "auto, (max-width: 42rem) 100vw, 40rem"

// photo is an image from a page bundle as it's published: resized into several
// widths in each of photoFormats plus a fallback, or, for formats that aren't
// resized, one copy as is.
type photo struct {
	Sources  []photoSet // in photoFormats order; empty when not resized
	Fallback photoSet   // JPEG or PNG, for the <img>
}

// photoSet is one format of a photo, at each width.
type photoSet struct {
	Type     string         // MIME type
	Variants []photoVariant // narrowest first
}

type photoVariant struct {
	URL           string
	Width, Height int
}

// Largest is the variant used for src="", the fallback for anything that ignores srcset.
func (s photoSet) Largest() photoVariant { return s.Variants[len(s.Variants)-1] }

// Srcset is the srcset="" attribute, or "" when there's only one file without a known width.
func (s photoSet) Srcset() string {
	if len(s.Variants) == 1 && s.Variants[0].Width == 0 {
		return ""
	}
	parts := make([]string, len(s.Variants))
	for i, v := range s.Variants {
		parts[i] = fmt.Sprintf("%s %dw", v.URL, v.Width)
	}
	return strings.Join(parts, ", ")
}

// photos resizes the images page bundles use. Resizing is slow, so each photo's
// variants are kept in cacheDir, keyed by a hash of the file and the settings
// above, and only made again when one of those changes. Originals are never
// published: re-encoding drops their EXIF data, which can hold where a photo was
// taken.
type photos struct {
	cacheDir string
	copies   map[string]string // output path (site-relative URL) → file to copy there
}

func newPhotos(cacheDir string) *photos {
	return &photos{cacheDir: cacheDir, copies: map[string]string{}}
}

// resizable formats are resized; anything else in a bundle (an SVG, a GIF) is published as is.
var resizable = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}

// add prepares src, a file in a page bundle, to be published under dirURL (the
// page's URL) and returns how to refer to it.
func (p *photos) add(src, dirURL string) (*photo, error) {
	ext := strings.ToLower(filepath.Ext(src))
	name := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	if !resizable[ext] {
		if _, err := os.Stat(src); err != nil {
			return nil, err
		}
		url := path.Join(dirURL, filepath.Base(src))
		p.copies[url] = src
		return &photo{Fallback: photoSet{Variants: []photoVariant{{URL: url}}}}, nil
	}

	dir, err := p.cached(src)
	if err != nil {
		return nil, fmt.Errorf("resizing %s: %w", filepath.Base(src), err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	sets := map[string]*photoSet{} // by extension
	for _, e := range entries {
		file := filepath.Join(dir, e.Name())
		cfg, err := decodeConfig(file)
		if err != nil {
			return nil, err
		}
		outExt := filepath.Ext(e.Name())
		url := path.Join(dirURL, fmt.Sprintf("%s.%dw%s", name, cfg.Width, outExt))
		p.copies[url] = file
		if sets[outExt] == nil {
			sets[outExt] = &photoSet{}
		}
		sets[outExt].Variants = append(sets[outExt].Variants, photoVariant{URL: url, Width: cfg.Width, Height: cfg.Height})
	}
	for _, s := range sets {
		slices.SortFunc(s.Variants, func(a, b photoVariant) int { return a.Width - b.Width })
	}

	ph := &photo{}
	for _, f := range photoFormats {
		if s := sets[f.ext]; s != nil {
			s.Type = f.mime
			ph.Sources = append(ph.Sources, *s)
		}
	}
	fallback := sets[fallbackExt(src)]
	if fallback == nil {
		return nil, fmt.Errorf("cache for %s has no %s files; delete %s", filepath.Base(src), fallbackExt(src), dir)
	}
	ph.Fallback = *fallback
	return ph, nil
}

// fallbackExt is the format the <img> uses: PNG for PNGs, so they keep any
// transparency, and JPEG for the rest.
func fallbackExt(src string) string {
	if strings.ToLower(filepath.Ext(src)) == ".png" {
		return ".png"
	}
	return ".jpg"
}

// cached returns the cache folder holding src's variants, one file per width
// and format named like "640.webp", making them first if they aren't there.
func (p *photos) cached(src string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(data)
	fmt.Fprint(h, photoWidths, jpegQuality, webpQuality, avifQuality, avifSpeed)
	dir := filepath.Join(p.cacheDir, hex.EncodeToString(h.Sum(nil))[:16])
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}

	// AutoOrientation applies the EXIF rotation phone cameras rely on, since the
	// re-encoded file won't carry it.
	img, err := imaging.Open(src, imaging.AutoOrientation(true))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(p.cacheDir, 0o755); err != nil {
		return "", err
	}
	// Written to a temporary folder and renamed into place, so a build that's
	// interrupted never leaves a half-filled entry that looks finished.
	tmp, err := os.MkdirTemp(p.cacheDir, "tmp-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	// Each width and format is encoded on its own goroutine, as many at a time
	// as there are CPUs: resizing is cheap next to encoding.
	widths := variantWidths(img.Bounds().Dx())
	resized := make([]image.Image, len(widths))
	for i, w := range widths {
		resized[i] = imaging.Resize(img, w, 0, imaging.Lanczos)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	slots := make(chan struct{}, runtime.NumCPU())
	save := func(name string, write func(io.Writer) error) {
		slots <- struct{}{}
		err := writeWith(filepath.Join(tmp, name), write)
		<-slots
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("%s: %w", name, err)
		}
	}
	for i, w := range widths {
		m := resized[i]
		base := strconv.Itoa(w)
		wg.Go(func() {
			save(base+fallbackExt(src), func(out io.Writer) error {
				format := imaging.JPEG
				if fallbackExt(src) == ".png" {
					format = imaging.PNG
				}
				return imaging.Encode(out, m, format, imaging.JPEGQuality(jpegQuality))
			})
		})
		for _, f := range photoFormats {
			wg.Go(func() {
				save(base+f.ext, func(out io.Writer) error { return f.encode(out, m) })
			})
		}
	}
	wg.Wait()
	if firstErr != nil {
		return "", firstErr
	}
	if err := os.Rename(tmp, dir); err != nil {
		// Fine if another build got there first.
		if _, statErr := os.Stat(dir); statErr != nil {
			return "", err
		}
	}
	return dir, nil
}

// writeWith creates path and fills it with write.
func writeWith(path string, write func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := write(f); err != nil {
		return err
	}
	return f.Close()
}

// variantWidths picks the photoWidths that fit a photo this wide, or just its
// own width when it's narrower than all of them.
func variantWidths(width int) []int {
	var widths []int
	for _, w := range photoWidths {
		if w <= width {
			widths = append(widths, w)
		}
	}
	if len(widths) == 0 {
		widths = []int{width}
	}
	return widths
}

func decodeConfig(file string) (image.Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return image.Config{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	return cfg, err
}

// writeTo copies every published photo into out.
func (p *photos) writeTo(out string) error {
	for url, src := range p.copies {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(out, filepath.FromSlash(url)), data); err != nil {
			return err
		}
	}
	return nil
}
