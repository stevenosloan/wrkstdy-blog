package site

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

// minifier shrinks CSS and JS before they're hashed, so the hash and the
// integrity attribute match the bytes that are served.
var minifier = func() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	return m
}()

var minifyTypes = map[string]string{".css": "text/css", ".js": "text/javascript"}

// Asset is a file from assets/ written under a content-hashed name, so browsers
// can cache it forever and still pick up changes.
type Asset struct {
	URL       string // like "/assets/css/main.3f2a9c1b04de.css"
	Integrity string // Subresource Integrity hash for the integrity="" attribute
}

// fingerprintDir writes every file under src into out/prefix under a hashed name,
// minifying CSS and JS on the way.
// CSS goes last, so its url("/prefix/...") references can be rewritten to the
// hashed names before the CSS itself is hashed.
func fingerprintDir(src, out, prefix string) (map[string]Asset, error) {
	var files, css []string
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == ".DS_Store" {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		rel = filepath.ToSlash(rel)
		if filepath.Ext(rel) == ".css" {
			css = append(css, rel)
		} else {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	assets := map[string]Asset{}
	urlRef := regexp.MustCompile(`url\("/` + regexp.QuoteMeta(prefix) + `/([^"]+)"\)`)
	for _, rel := range append(files, css...) {
		data, err := os.ReadFile(filepath.Join(src, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		if filepath.Ext(rel) == ".css" {
			var missing string
			data = urlRef.ReplaceAllFunc(data, func(m []byte) []byte {
				ref := string(urlRef.FindSubmatch(m)[1])
				a, ok := assets[ref]
				if !ok {
					missing = ref
					return m
				}
				return []byte(`url("` + a.URL + `")`)
			})
			if missing != "" {
				return nil, fmt.Errorf("%s: no asset %q in assets/", rel, missing)
			}
		}
		ext := filepath.Ext(rel)
		if mediatype, ok := minifyTypes[ext]; ok {
			if data, err = minifier.Bytes(mediatype, data); err != nil {
				return nil, fmt.Errorf("%s: %w", rel, err)
			}
		}

		sum := sha256.Sum256(data)
		name := prefix + "/" + strings.TrimSuffix(rel, ext) + "." + hex.EncodeToString(sum[:])[:12] + ext
		if err := writeFile(filepath.Join(out, filepath.FromSlash(name)), data); err != nil {
			return nil, err
		}
		assets[rel] = Asset{
			URL:       "/" + name,
			Integrity: "sha256-" + base64.StdEncoding.EncodeToString(sum[:]),
		}
	}
	return assets, nil
}

// copyDir copies every file under src into out, keeping relative paths.
func copyDir(src, out string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == ".DS_Store" {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(out, rel), data)
	})
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
