package store

import (
	"encoding/xml"
	"strings"

	"dev.jevido/jevidocs/services/api/app/models"
)

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// Sitemap lists every published page of a project for search engines.
func Sitemap(p models.Project) (string, error) {
	pages, err := pagesOf(p, true)
	if err != nil {
		return "", err
	}
	set := urlSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, pg := range pages {
		u := sitemapURL{Loc: PageURL(p, pg.Slug)}
		if pg.UpdatedAt != nil {
			u.LastMod = pg.UpdatedAt.ToDateString()
		}
		set.URLs = append(set.URLs, u)
	}
	out, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + strings.TrimSpace(string(out)) + "\n", nil
}
