package site

import (
	"encoding/xml"
	"path/filepath"
	"time"
)

// sitemapURL is where the sitemap is written, and what robots.txt points at.
const sitemapURL = "/sitemap.xml"

type urlset struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapLoc `xml:"url"`
}

type sitemapLoc struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

// writeSitemap lists every indexable page at sitemapURL. The 404 page is
// noindex, so it's left out, as are drafts from a dev build.
func writeSitemap(out string, site *Site, c *content) error {
	var locs []sitemapLoc
	add := func(p *Page, lastmod time.Time) {
		if p == nil || p.Draft {
			return
		}
		loc := sitemapLoc{Loc: site.AbsURL(p.URL)}
		if !lastmod.IsZero() {
			loc.Lastmod = lastmod.Format(time.DateOnly)
		}
		locs = append(locs, loc)
	}

	// Indexes use their newest post's date, like the feed's lastBuildDate,
	// so rebuilding without a new post doesn't report a change.
	var newest time.Time
	for _, s := range site.Sections() {
		if p := newestPost(s); p != nil && p.Date.After(newest) {
			newest = p.Date
		}
	}
	add(c.home, newest)
	for _, p := range c.pages {
		add(p, p.Date)
	}
	for _, s := range site.Sections() {
		var lastmod time.Time
		if p := newestPost(s); p != nil {
			lastmod = p.Date
		}
		add(s.Index, lastmod)
		for _, p := range s.Pages {
			add(p, p.Date)
		}
	}

	data, err := xml.MarshalIndent(urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  locs,
	}, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(out, filepath.FromSlash(sitemapURL)), append([]byte(xml.Header), data...))
}

// newestPost is a section's latest published post, skipping drafts so a dev
// build doesn't date an index by a post the sitemap leaves out.
func newestPost(s *Section) *Page {
	for _, p := range s.Pages {
		if !p.Draft {
			return p
		}
	}
	return nil
}
