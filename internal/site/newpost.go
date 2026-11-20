package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const starterFrontMatter = `---
title: %q
description: ""
draft: true
---

`

// starterBodies are what a new post in these sections starts with, after the
// front matter. Each has to build as it is, since drafts build with
// serve -drafts: a recipe block can't be empty, and a photo has to exist.
var starterBodies = map[string]string{
	"recipes": "## The recipe\n\n" +
		"```recipe\n" +
		"the finished dish\n" +
		"  the last step\n" +
		"    an earlier step\n" +
		"      an ingredient\n" +
		"      another ingredient\n" +
		"    an ingredient added at the end\n" +
		"```\n\n" +
		"## The process\n\n",
	"journal": "<!--\n" +
		"Import photos with: go tool mage photos <this folder> <photo or folder>\n" +
		"Then add one with its caption on the next line:\n\n" +
		"![What's in the photo](photo.jpg)\n" +
		"The caption.\n" +
		"-->\n\n",
}

// NewPost creates <root>/content/<section>/<date>-<slug>/index.md as a draft
// with starter front matter, deriving the slug from the title. Every post is a
// folder, so its photos can sit beside it. The date in the folder's name is the
// post's date. Recipes and journal entries start with a body to fill in. It
// returns the new file's path and never overwrites an existing
// post.
func NewPost(root, section, title string, date time.Time) (string, error) {
	title = strings.TrimSpace(title)
	slug := Slugify(title)
	if slug == "" {
		return "", fmt.Errorf("can't make a slug from title %q", title)
	}
	dir := filepath.Join(root, "content", filepath.FromSlash(section))
	name := date.Format(time.DateOnly) + "-" + slug
	path := filepath.Join(dir, name, "index.md")
	// The build still reads a post that's a plain file, and one with the same
	// name would publish to the same URL.
	if flat := filepath.Join(dir, name+".md"); fileExists(flat) {
		return "", fmt.Errorf("%s already exists", flat)
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

	body := fmt.Sprintf(starterFrontMatter, title) + starterBodies[filepath.Base(dir)]
	if _, err := f.WriteString(body); err != nil {
		return "", err
	}
	return path, f.Close()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
