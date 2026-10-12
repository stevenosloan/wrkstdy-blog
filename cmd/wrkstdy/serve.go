package main

import (
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"wrkstdy.com/site/internal/site"
)

// watched are the folders whose changes trigger a rebuild.
var watched = []string{"content", "templates", "static", "assets", "data"}

func runServe(args []string) error {
	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := flags.String("addr", "localhost:1313", "address to listen on")
	drafts := flags.Bool("drafts", false, "include draft pages")
	flags.Parse(args)

	opts := site.Options{Root: ".", Out: "public", Drafts: *drafts, BaseURL: "http://" + *addr + "/"}
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

	log.Printf("serving on http://%s/ (ctrl-c to stop)", *addr)
	return http.ListenAndServe(*addr, http.FileServer(http.Dir(opts.Out)))
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
					log.Printf("build failed: %v", err)
				}
			})
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("watch error: %v", err)
		}
	}
}
