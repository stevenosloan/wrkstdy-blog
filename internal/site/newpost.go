package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const postTemplate = `---
title: %q
description: ""
draft: true
---

`

// NewPost creates <root>/content/<section>/<date>-<slug>.md as a draft with
// starter front matter, deriving the slug from the title. The date in the
// filename is the post's date. With bundle, it makes <date>-<slug>/index.md
// instead, a folder to put the post's photos in. It returns the new file's path
// and never overwrites an existing post.
func NewPost(root, section, title string, date time.Time, bundle bool) (string, error) {
	title = strings.TrimSpace(title)
	slug := Slugify(title)
	if slug == "" {
		return "", fmt.Errorf("can't make a slug from title %q", title)
	}
	dir := filepath.Join(root, "content", filepath.FromSlash(section))
	name := date.Format(time.DateOnly) + "-" + slug
	path, other := filepath.Join(dir, name+".md"), filepath.Join(dir, name, "index.md")
	if bundle {
		path, other = other, path
	}
	// A bundle and a plain file with the same name would publish to the same URL.
	if _, err := os.Stat(other); err == nil {
		return "", fmt.Errorf("%s already exists", other)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	// O_EXCL makes this fail instead of overwriting an existing post.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, postTemplate, title); err != nil {
		return "", err
	}
	return path, f.Close()
}

// Slugify turns "Building a blog, on purpose" into "building-a-blog-on-purpose".
// Apostrophes are dropped so "I'm" becomes "im" rather than "i-m".
func Slugify(title string) string {
	var b strings.Builder
	pendingDash := false
	for _, r := range strings.ToLower(title) {
		switch {
		case r == '\'' || r == '’':
			continue
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if pendingDash && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingDash = false
			b.WriteRune(r)
		default:
			pendingDash = true
		}
	}
	return b.String()
}
