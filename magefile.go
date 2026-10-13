//go:build mage

// Tasks for working on the site. Run `go tool mage` to list them.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"wrkstdy.com/site/internal/site"
)

const serveAddr = "localhost:1313"

// watched are the folders whose changes trigger a rebuild in Serve.
var watched = []string{"content", "templates", "static", "assets", "data"}

// Serve runs the dev server on localhost:1313 and rebuilds on save.
func Serve(
	drafts *bool, // include posts with draft: true
) error {
	opts := site.Options{Root: ".", Out: "public", Drafts: drafts != nil && *drafts, BaseURL: "http://" + serveAddr + "/"}
	if err := build(opts); err != nil {
		return err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()
	if err := w.Add("site.toml"); err != nil {
		return err
	}
	for _, dir := range watched {
		if err := watchTree(w, dir); err != nil {
			return err
		}
	}
	go rebuildOnChange(w, opts)

	fmt.Printf("serving on http://%s/ (ctrl-c to stop)\n", serveAddr)
	return http.ListenAndServe(serveAddr, http.FileServer(http.Dir(opts.Out)))
}

// Build writes the site to public/. Netlify runs it with -production on every deploy.
func Build(
	production *bool, // run go vet and the tests first, and fail the build if they fail
	baseURL *string, // override site.toml's baseURL, e.g. for deploy previews
) error {
	if production != nil && *production {
		mg.Deps(Test)
	}
	opts := site.Options{Root: ".", Out: "public"}
	if baseURL != nil {
		opts.BaseURL = *baseURL
	}
	return build(opts)
}

// Test runs go vet and the tests.
func Test() error {
	if err := sh.RunV("go", "vet", "./..."); err != nil {
		return err
	}
	return sh.RunV("go", "test", "./...")
}

// New asks for a title and creates a draft post slugged from it.
func New(
	section *string, // folder under content/ to create the post in (default blog)
) error {
	fmt.Print("Title: ")
	title, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && title == "" {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("a title is required")
	}
	dir := "blog"
	if section != nil && *section != "" {
		dir = *section
	}
	path, err := site.NewPost(".", dir, title)
	if err != nil {
		return err
	}
	fmt.Printf("created %s\n", path)
	return nil
}

func build(opts site.Options) error {
	start := time.Now()
	n, err := site.Build(opts)
	if err != nil {
		return err
	}
	fmt.Printf("built %d pages into %s/ in %s\n", n, opts.Out, time.Since(start).Round(time.Millisecond))
	return nil
}

// watchTree watches dir and every folder under it; fsnotify isn't recursive.
func watchTree(w *fsnotify.Watcher, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		return w.Add(path)
	})
}

// rebuildOnChange rebuilds once things go quiet, since one save can fire several events.
func rebuildOnChange(w *fsnotify.Watcher, opts site.Options) {
	var timer *time.Timer
	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Create) {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					watchTree(w, ev.Name)
				}
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(100*time.Millisecond, func() {
				if err := build(opts); err != nil {
					fmt.Printf("build failed: %v\n", err)
				}
			})
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			fmt.Printf("watch error: %v\n", err)
		}
	}
}
