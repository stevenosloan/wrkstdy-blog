package site

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // registers JPEG with image.DecodeConfig
	_ "image/png"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif" // also registers AVIF and WebP with image.DecodeConfig
	"github.com/gen2brain/webp"
)

// photoWidths are the widths a photo is resized to. The content column is 40rem
// (640px), so these cover 1x and 2x screens, and phones at 3x, which are
// narrower than the column. A 1920px copy would only go to 3x screens at
// desktop width, and cost as much as the rest together. Widths wider than the
// photo are skipped, so nothing is upscaled.
var photoWidths = []int{640, 960, 1280}

// colorVersion is part of each cached photo's key; bump it when openPhoto
// changes how pixels come out, so the cache is made again.
const colorVersion = "srgb-1"

const (
	jpegQuality = 80
	webpQuality = 70
	avifQuality = 50
	avifSpeed   = 6 // 0 (slowest, smallest) to 10; slower saves a few percent for several times the time
	webpMethod  = 6 // 0 (fastest) to 6; the slowest is still quick, and a few percent smaller
)

// avifChroma halves the color resolution, as photo encoders normally do: it's
// smaller and faster to encode than full color (4:4:4), which is what
// gen2brain/avif falls back to when Options leaves it unset.
const avifChroma = image.YCbCrSubsampleRatio420

// photoFormats are the formats each width is also encoded in, best first, as
// <source>s in a <picture>. The browser takes the first it supports, and falls
// back to the <img>, a JPEG (or a PNG, for PNGs, to keep transparency).
var photoFormats = []struct {
	ext, mime string
	encode    func(io.Writer, image.Image) error
}{
	{".avif", "image/avif", func(w io.Writer, m image.Image) error {
		return avif.Encode(w, m, avif.Options{Quality: avifQuality, QualityAlpha: avifQuality, Speed: avifSpeed, ChromaSubsampling: avifChroma})
	}},
	{".webp", "image/webp", func(w io.Writer, m image.Image) error {
		return webp.Encode(w, m, webp.Options{Quality: webpQuality, Method: webpMethod})
	}},
}

// photoSizes is the sizes="" attribute for a resized photo. "auto" lets the
// browser pick from the width the image is laid out at; browsers that don't know
// it yet fall back to the rest. It only works with loading="lazy", so a photo
// marked {loading=eager} gets photoSizesEager.
const (
	photoSizesEager = "(max-width: 42rem) 100vw, 40rem"
	photoSizes      = "auto, " + photoSizesEager
)

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

// Srcset is the srcset="" attribute, or "" when there's only one file to pick from.
func (s photoSet) Srcset() string {
	if len(s.Variants) < 2 {
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
//
// Published photos go under photosURL, mirroring the page's path, with a hash
// of their content in the name, like
// /assets/photos/recipes/ssamjang-wings/seasoned.1280w.3f2a9c1b04de.avif. A
// changed photo gets a new name, so like everything else under /assets/, they
// can be cached forever.
type photos struct {
	cacheDir string
	copies   map[string]string // output path (site-relative URL) → file to copy there
}

const photosURL = "/assets/photos"

// publish arranges for file to be copied into the site as name (like
// "seasoned.1280w.avif") under the page at dirURL, and returns its URL, with a
// hash of its content added before the extension.
func (p *photos) publish(file, dirURL, name string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	ext := filepath.Ext(name)
	url := path.Join(photosURL, dirURL, strings.TrimSuffix(name, ext)+"."+hex.EncodeToString(sum[:])[:12]+ext)
	p.copies[url] = file
	return url, nil
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
		url, err := p.publish(src, dirURL, filepath.Base(src))
		if err != nil {
			return nil, err
		}
		v := photoVariant{URL: url}
		// Its size, when it's a format Go can read (WebP and AVIF, not SVG), so
		// the page can save room for it before it loads.
		if cfg, err := decodeConfig(src); err == nil {
			v.Width, v.Height = cfg.Width, cfg.Height
		}
		return &photo{Fallback: photoSet{Variants: []photoVariant{v}}}, nil
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
		url, err := p.publish(file, dirURL, fmt.Sprintf("%s.%dw%s", name, cfg.Width, outExt))
		if err != nil {
			return nil, err
		}
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
	fmt.Fprint(h, photoWidths, jpegQuality, webpQuality, webpMethod, avifQuality, avifSpeed, avifChroma, colorVersion)
	dir := filepath.Join(p.cacheDir, hex.EncodeToString(h.Sum(nil))[:16])
	if _, err := os.Stat(dir); err == nil {
		// Marks it as used, so prune keeps it.
		now := time.Now()
		return dir, os.Chtimes(dir, now, now)
	}

	img, err := openPhoto(src)
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

// cacheMaxAge is how long a cached photo is kept after a build last used it.
// It's long enough that a photo in a draft, or in a post that was taken down
// for a while, doesn't have to be encoded again; anything older is likely a
// photo that's since been edited or replaced.
const cacheMaxAge = 30 * 24 * time.Hour

// prune deletes cached photos no build has used in cacheMaxAge, and temporary
// folders left by an interrupted build. Without it, the cache only grows, and
// on Netlify, the whole cache is restored and saved on every deploy.
func (p *photos) prune(now time.Time) error {
	entries, err := os.ReadDir(p.cacheDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			return err
		}
		age := now.Sub(info.ModTime())
		stale := age > cacheMaxAge
		if strings.HasPrefix(e.Name(), "tmp-") {
			stale = age > time.Hour
		}
		if stale {
			if err := os.RemoveAll(filepath.Join(p.cacheDir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
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
