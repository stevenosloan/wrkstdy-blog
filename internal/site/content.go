package site

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Page is one rendered markdown file.
type Page struct {
	Title       string
	Description string
	Date        time.Time
	Draft       bool

	URL        string // site-relative, like "/blog/hello-world/"
	Content    template.HTML
	HasCode    bool // loads MicroLighter
	HasMermaid bool // loads mermaid.js
}

// Section is a content directory with an _index.md, like content/blog/.
type Section struct {
	Name  string
	Index *Page
	Pages []*Page // newest first
}

type frontMatter struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	Draft       bool      `yaml:"draft"`
	Slug        string    `yaml:"slug"`
}

// splitFrontMatter separates a leading "---" YAML block from the markdown body.
func splitFrontMatter(src []byte) (meta, body []byte, err error) {
	src = bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	rest, ok := bytes.CutPrefix(src, []byte("---\n"))
	if !ok {
		return nil, src, nil
	}
	if !bytes.HasSuffix(rest, []byte("\n")) {
		rest = append(rest, '\n')
	}
	meta, body, ok = bytes.Cut(rest, []byte("\n---\n"))
	if !ok {
		return nil, nil, errors.New("front matter has no closing ---")
	}
	return meta, body, nil
}

// notFoundURL is where Netlify looks for the page it serves for missing paths.
const notFoundURL = "/404.html"

// content is everything loaded from the content directory.
type content struct {
	home     *Page
	sections map[string]*Section
	pages    []*Page // pages outside any section, like content/about.md at /about/
	notFound *Page   // content/404.md, written to /404.html for Netlify to serve on misses
}

func (b *builder) loadContent(dir string) (*content, error) {
	c := &content{sections: map[string]*Section{}}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		name, folder := filepath.Base(rel), filepath.Dir(rel)

		// A page bundle: content/journal/2026-10-15-lake-day/index.md is a page
		// with its photos beside it. The folder's name stands in for the filename.
		bundle := ""
		if name == "index.md" && folder != "." {
			bundle = filepath.Dir(path)
			name = filepath.Base(folder) + ".md"
			folder = filepath.Dir(folder)
		}

		section := filepath.ToSlash(folder)
		if section == "." {
			section = ""
		}
		urlFor := func(slug string) string {
			switch {
			case name == "_index.md" && section == "":
				return "/"
			case name == "_index.md":
				return "/" + section + "/"
			case rel == "404.md":
				return notFoundURL
			case section == "":
				return "/" + slug + "/"
			}
			return "/" + section + "/" + slug + "/"
		}

		page, err := b.loadPage(path, name, bundle, urlFor)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		switch {
		case page == nil: // a draft, left out
		case name == "_index.md" && section == "":
			c.home = page
		case name == "_index.md":
			c.section(section).Index = page
		case rel == "404.md":
			c.notFound = page
		case section == "":
			c.pages = append(c.pages, page)
		default:
			s := c.section(section)
			s.Pages = append(s.Pages, page)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := c.checkURLs(); err != nil {
		return nil, err
	}
	if c.home == nil {
		return nil, errors.New("content/_index.md is missing")
	}

	for name, s := range c.sections {
		if s.Index == nil {
			return nil, fmt.Errorf("content/%s/_index.md is missing", name)
		}
		sort.SliceStable(s.Pages, func(i, j int) bool {
			return s.Pages[i].Date.After(s.Pages[j].Date)
		})
	}
	return c, nil
}

// splitDatePrefix splits "2026-10-12-my-post" into its date and "my-post".
// Names without a leading date come back unchanged with a zero date.
func splitDatePrefix(name string) (time.Time, string) {
	const layout = "2006-01-02"
	if len(name) <= len(layout)+1 || name[len(layout)] != '-' {
		return time.Time{}, name
	}
	date, err := time.Parse(layout, name[:len(layout)])
	if err != nil {
		return time.Time{}, name
	}
	return date, name[len(layout)+1:]
}

// checkURLs fails when two pages would be written to the same URL, such as
// posts with the same slug on different dates.
func (c *content) checkURLs() error {
	seen := map[string]bool{}
	pages := append([]*Page{}, c.pages...)
	for _, s := range c.sections {
		pages = append(pages, s.Pages...)
	}
	for _, p := range pages {
		if seen[p.URL] {
			return fmt.Errorf("two pages share the URL %s; give one a different title or slug", p.URL)
		}
		seen[p.URL] = true
	}
	return nil
}

func (c *content) section(name string) *Section {
	s, ok := c.sections[name]
	if !ok {
		s = &Section{Name: name}
		c.sections[name] = s
	}
	return s
}

// loadPage reads and renders the markdown file at path, or returns nil for a
// draft when drafts are left out. name is the file's name, or for a bundle, its
// folder's name plus ".md"; bundle is that folder. urlFor maps the page's slug
// to its URL.
func (b *builder) loadPage(path, name, bundle string, urlFor func(slug string) string) (*Page, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	meta, body, err := splitFrontMatter(src)
	if err != nil {
		return nil, err
	}
	var fm frontMatter
	if err := yaml.Unmarshal(meta, &fm); err != nil {
		return nil, fmt.Errorf("front matter: %w", err)
	}
	if fm.Draft && !b.opts.Drafts {
		return nil, nil
	}

	fileDate, slug := splitDatePrefix(strings.TrimSuffix(name, ".md"))
	if fm.Slug != "" {
		slug = fm.Slug
	}
	date := fm.Date
	if date.IsZero() {
		date = fileDate
	}
	url := urlFor(slug)

	html, err := b.renderMarkdown(body, &pageInfo{URL: url, Bundle: bundle, photos: b.photos})
	if err != nil {
		return nil, err
	}
	return &Page{
		Title:       fm.Title,
		Description: fm.Description,
		Date:        date,
		Draft:       fm.Draft,
		URL:         url,
		Content:     template.HTML(html),
		HasCode:     strings.Contains(html, `<pre><code class="language-`),
		HasMermaid:  strings.Contains(html, `<pre class="mermaid">`),
	}, nil
}
