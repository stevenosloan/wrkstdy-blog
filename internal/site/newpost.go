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
date: %s
description: ""
draft: true
---

`

// NewPost creates <root>/content/<section>/<slug>.md as a draft with starter
// front matter, deriving the slug from the title. It returns the new file's path
// and never overwrites an existing post.
func NewPost(root, section, title string) (string, error) {
	title = strings.TrimSpace(title)
	slug := Slugify(title)
	if slug == "" {
		return "", fmt.Errorf("can't make a slug from title %q", title)
	}
	path := filepath.Join(root, "content", filepath.FromSlash(section), slug+".md")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	// O_EXCL makes this fail instead of overwriting an existing post.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, postTemplate, title, time.Now().Format("2006-01-02")); err != nil {
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
