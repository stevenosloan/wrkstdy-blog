package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/disintegration/imaging"
)

// importMaxSize is the longest side of an imported photo. The build makes
// copies up to 1280px wide from it, so this leaves room to crop, or to add
// wider copies later.
const importMaxSize = 2560

const importJPEGQuality = 90

// ImportedPhoto is one photo ImportPhotos wrote.
type ImportedPhoto struct {
	Path          string
	Width, Height int
}

// ImportPhotos copies photos into a post's folder (a page bundle, holding
// index.md) ready to commit: upright, converted to sRGB, no larger than
// importMaxSize, and without metadata, so no location ends up in the repo.
//
// from is a photo or a folder of them. Each keeps its name, lowercased, with
// .jpg for JPEGs. It won't overwrite a file in the post unless replace is set,
// and checks every name before writing anything.
func ImportPhotos(post, from string, replace bool) ([]ImportedPhoto, error) {
	if _, err := os.Stat(filepath.Join(post, "index.md")); err != nil {
		return nil, fmt.Errorf("%s isn't a post folder with an index.md; make one with go tool mage new -bundle", post)
	}

	srcs, err := photoFiles(from)
	if err != nil {
		return nil, err
	}
	if len(srcs) == 0 {
		return nil, fmt.Errorf("no JPEG or PNG photos in %s", from)
	}

	dsts := make([]string, len(srcs))
	var clashes []string
	for i, src := range srcs {
		ext := ".jpg"
		if strings.EqualFold(filepath.Ext(src), ".png") {
			ext = ".png"
		}
		name := strings.ToLower(strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))) + ext
		dsts[i] = filepath.Join(post, name)
		if slices.Contains(dsts[:i], dsts[i]) {
			return nil, fmt.Errorf("two photos would both be named %s", name)
		}
		if _, err := os.Stat(dsts[i]); err == nil && !replace {
			clashes = append(clashes, name)
		}
	}
	if len(clashes) > 0 {
		return nil, fmt.Errorf("%s already in %s; pass -replace to overwrite", strings.Join(clashes, ", "), post)
	}

	var done []ImportedPhoto
	for i, src := range srcs {
		img, err := openPhoto(src)
		if err != nil {
			return done, fmt.Errorf("%s: %w", filepath.Base(src), err)
		}
		fitted := imaging.Fit(img, importMaxSize, importMaxSize, imaging.Lanczos)
		// imaging writes no EXIF, ICC or other metadata, only the pixels.
		if err := imaging.Save(fitted, dsts[i], imaging.JPEGQuality(importJPEGQuality)); err != nil {
			return done, err
		}
		done = append(done, ImportedPhoto{Path: dsts[i], Width: fitted.Bounds().Dx(), Height: fitted.Bounds().Dy()})
	}
	return done, nil
}

// photoFiles lists from if it's a photo, or the photos directly inside it if
// it's a folder, by name.
func photoFiles(from string) ([]string, error) {
	info, err := os.Stat(from)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !resizable[strings.ToLower(filepath.Ext(from))] {
			return nil, unsupportedPhoto(from)
		}
		return []string{from}, nil
	}

	entries, err := os.ReadDir(from)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		switch {
		case e.IsDir() || strings.HasPrefix(e.Name(), "."):
		case resizable[ext]:
			files = append(files, filepath.Join(from, e.Name()))
		case ext == ".heic" || ext == ".heif":
			return nil, unsupportedPhoto(e.Name())
		}
	}
	return files, nil
}

func unsupportedPhoto(name string) error {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".heic" || ext == ".heif" {
		return fmt.Errorf("%s is HEIC, which can't be read yet; export it as JPEG", filepath.Base(name))
	}
	return errors.New(filepath.Base(name) + " isn't a JPEG or PNG")
}
