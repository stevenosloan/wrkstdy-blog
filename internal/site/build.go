// Package site builds wrkstdy.com from markdown into a folder of static files.
package site

import (
	"errors"
	"fmt"
	"html/template"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

type Options struct {
	Root    string // project folder holding site.toml
	Out     string // output folder, wiped and rewritten on every build
	Drafts  bool   // include pages with draft: true
	BaseURL string // overrides baseURL from site.toml, for deploy previews and local serving

	// Bundler builds the CSS and JS. Leave it nil for a one-off production build;
	// a dev server passes one it keeps, so rebuilds are incremental.
	Bundler *Bundler
}

// Site is what templates see as .Site.
type Site struct {
	Config
	Projects []Project
	Now      time.Time

	sections map[string]*Section
	assets   map[string]Asset
}

// Section returns the named content section, or nil if there isn't one.
func (s *Site) Section(name string) *Section {
	return s.sections[name]
}

// Sections returns every content section, by name.
func (s *Site) Sections() []*Section {
	var list []*Section
	for _, name := range slices.Sorted(maps.Keys(s.sections)) {
		list = append(list, s.sections[name])
	}
	return list
}

// Asset returns the fingerprinted URL and integrity hash for a file in assets/.
func (s *Site) Asset(path string) (Asset, error) {
	a, ok := s.assets[path]
	if !ok {
		return Asset{}, fmt.Errorf("no asset %q in assets/", path)
	}
	return a, nil
}

// AbsURL turns a site-relative path into a full URL.
func (s *Site) AbsURL(path string) string {
	return strings.TrimSuffix(s.BaseURL, "/") + path
}

// view is the data passed to every page template.
type view struct {
	Site    *Site
	Page    *Page
	Section *Section // set on a section's index page
}

type builder struct {
	opts       Options
	md         goldmark.Markdown
	shortcodes *template.Template
	pages      map[string]*template.Template // by kind: home, archive, single
	photos     *photos                       // images from page bundles, resized
}

// Build renders the whole site into opts.Out and returns how many pages it wrote.
func Build(opts Options) (int, error) {
	root := opts.Root
	if err := checkOut(root, opts.Out); err != nil {
		return 0, err
	}

	cfg, err := loadConfig(filepath.Join(root, "site.toml"))
	if err != nil {
		return 0, err
	}
	if opts.BaseURL != "" {
		cfg.BaseURL = opts.BaseURL
	}
	projects, err := loadProjects(filepath.Join(root, "data", "projects.toml"))
	if err != nil {
		return 0, err
	}

	b := &builder{opts: opts, md: newMarkdown(), photos: newPhotos(filepath.Join(root, ".cache", "images"))}
	if err := b.loadTemplates(filepath.Join(root, "templates")); err != nil {
		return 0, err
	}

	c, err := b.loadContent(filepath.Join(root, "content"))
	if err != nil {
		return 0, err
	}

	if err := os.RemoveAll(opts.Out); err != nil {
		return 0, err
	}
	if err := copyDir(filepath.Join(root, "static"), opts.Out); err != nil {
		return 0, fmt.Errorf("copying static: %w", err)
	}
	if err := b.photos.writeTo(opts.Out); err != nil {
		return 0, fmt.Errorf("copying photos: %w", err)
	}
	assets, err := fingerprintDir(filepath.Join(root, "assets"), opts.Out, "assets")
	if err != nil {
		return 0, fmt.Errorf("fingerprinting assets: %w", err)
	}
	bundler := opts.Bundler
	if bundler == nil {
		bundler = NewBundler(root, opts.Out, false)
		defer bundler.Close()
	}
	built, err := bundler.Build()
	if err != nil {
		return 0, fmt.Errorf("bundling assets: %w", err)
	}
	maps.Copy(assets, built)

	site := &Site{
		Config:   cfg,
		Projects: projects,
		Now:      time.Now(),
		sections: c.sections,
		assets:   assets,
	}

	count := 0
	render := func(kind string, v view) error {
		v.Site = site
		count++
		return b.renderPage(kind, v)
	}

	if err := render("home", view{Page: c.home}); err != nil {
		return 0, err
	}
	if c.notFound != nil {
		if err := render("single", view{Page: c.notFound}); err != nil {
			return 0, err
		}
	}
	for _, p := range c.pages {
		if err := render("single", view{Page: p}); err != nil {
			return 0, err
		}
	}
	for _, s := range c.sections {
		if err := render("archive", view{Page: s.Index, Section: s}); err != nil {
			return 0, err
		}
		if err := writeFeed(opts.Out, site, s); err != nil {
			return 0, fmt.Errorf("writing %s: %w", s.FeedURL(), err)
		}
		for _, p := range s.Pages {
			if err := render("single", view{Page: p}); err != nil {
				return 0, err
			}
		}
	}
	return count, nil
}

func (b *builder) renderPage(kind string, v view) error {
	path := filepath.Join(b.opts.Out, filepath.FromSlash(v.Page.URL))
	if !strings.HasSuffix(v.Page.URL, ".html") {
		path = filepath.Join(path, "index.html")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := b.pages[kind].ExecuteTemplate(f, "base.html", v); err != nil {
		return fmt.Errorf("rendering %s: %w", v.Page.URL, err)
	}
	return f.Close()
}

// checkOut refuses output folders that would wipe the project when cleaned.
func checkOut(root, out string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absOut, absRoot)
	if err == nil && !strings.HasPrefix(rel, "..") {
		return errors.New("the output folder can't be the project folder or one of its parents")
	}
	return nil
}
