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

// Page is one rendered markdown file, or a generated page like a tag listing.
type Page struct {
	Title       string
	Description string
	Date        time.Time
	Tags        []string
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

// Recent returns up to n of the newest pages.
func (s *Section) Recent(n int) []*Page {
	return s.Pages[:min(n, len(s.Pages))]
}

type YearGroup struct {
	Year  int
	Pages []*Page
}

// ByYear groups pages by year, newest first.
func (s *Section) ByYear() []YearGroup {
	var groups []YearGroup
	for _, p := range s.Pages {
		y := p.Date.Year()
		if len(groups) == 0 || groups[len(groups)-1].Year != y {
			groups = append(groups, YearGroup{Year: y})
		}
		last := &groups[len(groups)-1]
		last.Pages = append(last.Pages, p)
	}
	return groups
}

type Tag struct {
	Name  string
	URL   string
	Pages []*Page
}

type frontMatter struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        time.Time `yaml:"date"`
	Tags        []string  `yaml:"tags"`
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

// content is everything loaded from the content directory.
type content struct {
	home     *Page
	sections map[string]*Section
	pages    []*Page // pages outside any section, like /styleguide/
}

func (b *builder) loadContent(dir string) (*content, error) {
	c := &content{sections: map[string]*Section{}}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		page, slug, err := b.loadPage(path)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if page.Draft && !b.opts.Drafts {
			return nil
		}

		section := filepath.ToSlash(filepath.Dir(rel))
		if section == "." {
			section = ""
		}
		if filepath.Base(rel) == "_index.md" {
			if section == "" {
				page.URL = "/"
				c.home = page
				return nil
			}
			page.URL = "/" + section + "/"
			c.section(section).Index = page
			return nil
		}

		if section == "" {
			page.URL = "/" + slug + "/"
			c.pages = append(c.pages, page)
			return nil
		}
		page.URL = "/" + section + "/" + slug + "/"
		s := c.section(section)
		s.Pages = append(s.Pages, page)
		return nil
	})
	if err != nil {
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

func (c *content) section(name string) *Section {
	s, ok := c.sections[name]
	if !ok {
		s = &Section{Name: name}
		c.sections[name] = s
	}
	return s
}

func (b *builder) loadPage(path string) (page *Page, slug string, err error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	meta, body, err := splitFrontMatter(src)
	if err != nil {
		return nil, "", err
	}
	var fm frontMatter
	if err := yaml.Unmarshal(meta, &fm); err != nil {
		return nil, "", fmt.Errorf("front matter: %w", err)
	}

	html, err := b.renderMarkdown(body)
	if err != nil {
		return nil, "", err
	}

	slug = fm.Slug
	if slug == "" {
		slug = strings.TrimSuffix(filepath.Base(path), ".md")
	}
	return &Page{
		Title:       fm.Title,
		Description: fm.Description,
		Date:        fm.Date,
		Tags:        fm.Tags,
		Draft:       fm.Draft,
		Content:     template.HTML(html),
		HasCode:     strings.Contains(html, `<pre><code class="language-`),
		HasMermaid:  strings.Contains(html, `<pre class="mermaid">`),
	}, slug, nil
}

// collectTags groups every sectioned page by tag, most-used tag first.
func collectTags(sections map[string]*Section) []*Tag {
	byName := map[string]*Tag{}
	for _, s := range sections {
		for _, p := range s.Pages {
			for _, name := range p.Tags {
				t, ok := byName[name]
				if !ok {
					t = &Tag{Name: name, URL: tagURL(name)}
					byName[name] = t
				}
				t.Pages = append(t.Pages, p)
			}
		}
	}

	tags := make([]*Tag, 0, len(byName))
	for _, t := range byName {
		sort.SliceStable(t.Pages, func(i, j int) bool {
			return t.Pages[i].Date.After(t.Pages[j].Date)
		})
		tags = append(tags, t)
	}
	sort.Slice(tags, func(i, j int) bool {
		if len(tags[i].Pages) != len(tags[j].Pages) {
			return len(tags[i].Pages) > len(tags[j].Pages)
		}
		return tags[i].Name < tags[j].Name
	})
	return tags
}

func tagURL(name string) string {
	return "/tags/" + strings.ReplaceAll(strings.ToLower(name), " ", "-") + "/"
}
