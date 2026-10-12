package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const postTemplate = `---
title: %q
date: %s
description: ""
tags: []
draft: true
---

`

// runNew creates content/<section>/<slug>.md with starter front matter.
func runNew(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: wrkstdy new blog/my-post")
	}
	rel := strings.TrimSuffix(args[0], ".md") + ".md"
	path := filepath.Join("content", filepath.FromSlash(rel))

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// O_EXCL makes this fail instead of overwriting an existing post.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	slug := strings.TrimSuffix(filepath.Base(path), ".md")
	if _, err := fmt.Fprintf(f, postTemplate, titleFromSlug(slug), time.Now().Format("2006-01-02")); err != nil {
		return err
	}
	log.Printf("created %s", path)
	return f.Close()
}

// titleFromSlug turns "my-first-post" into "My first post".
func titleFromSlug(slug string) string {
	s := strings.ReplaceAll(slug, "-", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
