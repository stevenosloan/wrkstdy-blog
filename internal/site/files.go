package site

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Asset is a file from assets/ written under a content-hashed name, so browsers
// can cache it forever and still pick up changes.
type Asset struct {
	URL       string // like "/assets/css/main.3f2a9c1b04de.css"
	Integrity string // Subresource Integrity hash for the integrity="" attribute
}

// fingerprintDir writes every file under src that esbuild doesn't build (see
// Bundler) into out/prefix under a hashed name: icons, and whatever else is
// dropped in assets/.
func fingerprintDir(src, out, prefix string) (map[string]Asset, error) {
	assets := map[string]Asset{}
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == ".DS_Store" {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		rel = filepath.ToSlash(rel)
		ext := filepath.Ext(rel)
		if ext == ".css" || ext == ".js" || ext == ".woff2" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		name := prefix + "/" + strings.TrimSuffix(rel, ext) + "." + hex.EncodeToString(sum[:])[:12] + ext
		if err := writeFile(filepath.Join(out, filepath.FromSlash(name)), data); err != nil {
			return err
		}
		assets[rel] = Asset{
			URL:       "/" + name,
			Integrity: "sha256-" + base64.StdEncoding.EncodeToString(sum[:]),
		}
		return nil
	})
	return assets, err
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
