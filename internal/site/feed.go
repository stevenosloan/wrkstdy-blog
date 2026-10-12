package site

import (
	"encoding/xml"
	"path/filepath"
	"time"
)

// RSS 2.0, shaped as Go structs so encoding/xml can write it.
type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	Items         []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

// writeFeed writes /<section>/index.xml with the section's newest pages.
func writeFeed(site *Site, s *Section, out string) error {
	ch := rssChannel{
		Title:       site.Title + " " + s.Index.Title,
		Link:        site.AbsURL(s.Index.URL),
		Description: s.Index.Description,
		Language:    site.Language,
	}
	for _, p := range s.Recent(site.FeedLimit) {
		link := site.AbsURL(p.URL)
		ch.Items = append(ch.Items, rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			PubDate:     p.Date.Format(time.RFC1123Z),
			Description: p.Description,
		})
	}
	if len(s.Pages) > 0 {
		ch.LastBuildDate = s.Pages[0].Date.Format(time.RFC1123Z)
	}

	data, err := xml.MarshalIndent(rss{Version: "2.0", Channel: ch}, "", "  ")
	if err != nil {
		return err
	}
	data = append([]byte(xml.Header), data...)
	return writeFile(filepath.Join(out, s.Name, "index.xml"), data)
}
