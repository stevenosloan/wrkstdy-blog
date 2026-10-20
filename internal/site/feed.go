package site

import (
	"encoding/xml"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FeedURL is where a section's RSS feed is written, like "/blog/feed.xml".
func (s *Section) FeedURL() string {
	return s.Index.URL + "feed.xml"
}

// FeedTitle names a section's feed in a reader, like "wrkstdy · journal".
func (s *Section) FeedTitle(site *Site) string {
	return site.Title + " · " + s.Index.Title
}

// Year is one year's worth of a section's posts, for the archive.
type Year struct {
	Year  int
	Pages []*Page // newest first
}

// Years groups a section's posts by year, newest first.
func (s *Section) Years() []Year {
	var years []Year
	for _, p := range s.Pages {
		if n := len(years); n == 0 || years[n-1].Year != p.Date.Year() {
			years = append(years, Year{Year: p.Date.Year()})
		}
		years[len(years)-1].Pages = append(years[len(years)-1].Pages, p)
	}
	return years
}

// RSS 2.0, with the content module for full post HTML and an atom:link
// pointing back at the feed, which validators ask for.
type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Content string     `xml:"xmlns:content,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	Self          atomLink  `xml:"atom:link"`
	Items         []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description,omitempty"`
	Content     string `xml:"content:encoded"`
}

// writeFeed writes the section's posts as RSS to its FeedURL.
func writeFeed(out string, site *Site, s *Section) error {
	ch := rssChannel{
		Title:       s.FeedTitle(site),
		Link:        site.AbsURL(s.Index.URL),
		Description: s.Index.Description,
		Language:    site.Language,
		Self:        atomLink{Href: site.AbsURL(s.FeedURL()), Rel: "self", Type: "application/rss+xml"},
	}
	if ch.Description == "" {
		ch.Description = site.Description
	}
	// The newest post's date rather than the build time, so rebuilding
	// without a new post doesn't tell readers something changed.
	if len(s.Pages) > 0 {
		ch.LastBuildDate = s.Pages[0].Date.Format(time.RFC1123Z)
	}
	for _, p := range s.Pages {
		link := site.AbsURL(p.URL)
		ch.Items = append(ch.Items, rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			PubDate:     p.Date.Format(time.RFC1123Z),
			Description: p.Description,
			Content:     absLinks(string(p.Content), site.AbsURL("")),
		})
	}

	data, err := xml.MarshalIndent(rss{
		Version: "2.0",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: ch,
	}, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(out, filepath.FromSlash(s.FeedURL())), append([]byte(xml.Header), data...))
}

// absLinks points site-relative links and images at the full URL, since a
// feed reader shows posts away from the site and may not resolve them.
func absLinks(html, base string) string {
	html = strings.NewReplacer(
		`href="/`, `href="`+base+`/`,
		`src="/`, `src="`+base+`/`,
	).Replace(html)
	// srcset lists several URLs, so each one needs the base.
	return srcsetRE.ReplaceAllStringFunc(html, func(attr string) string {
		return srcsetURLRE.ReplaceAllString(attr, "${1}"+base+"/")
	})
}

var (
	srcsetRE    = regexp.MustCompile(`srcset="[^"]*"`)
	srcsetURLRE = regexp.MustCompile(`(srcset="|,\s*)/`)
)
