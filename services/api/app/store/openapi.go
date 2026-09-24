package store

import (
	"strings"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/openapi"
)

// ImportOpenAPI generates an API reference from an OpenAPI 3 document into
// pages under prefix: new pages are created, changed ones updated, and pages
// under prefix the spec no longer produces are deleted. Pages outside prefix
// are never touched.
func ImportOpenAPI(p models.Project, spec []byte, prefix string) (SyncResult, error) {
	var res SyncResult
	prefix = docs.NormalizeSlug(prefix)
	if prefix == "" {
		prefix = openapi.DefaultPrefix
	}
	if !pageSlugRe.MatchString(prefix) {
		return res, ValidationError{"prefix must be a path of lowercase segments, like api-reference"}
	}
	base := strings.TrimPrefix(PageURL(p, ""), SiteURL())
	pages, err := openapi.Generate(spec, prefix, base)
	if err != nil {
		return res, ValidationError{err.Error()}
	}

	existing, err := AdminPages(p)
	if err != nil {
		return res, err
	}
	bySlug := map[string]models.Page{}
	for _, pg := range existing {
		if pg.Slug == prefix || strings.HasPrefix(pg.Slug, prefix+"/") {
			bySlug[pg.Slug] = pg
		}
	}
	seen := map[string]bool{}
	pub := true
	for _, g := range pages {
		seen[g.Slug] = true
		in := PageInput{Slug: g.Slug, Title: g.Title, Description: g.Description, Position: g.Position, Body: g.Body, Published: &pub}
		cur, ok := bySlug[g.Slug]
		if !ok {
			if _, err := SavePage(p, in, nil); err != nil {
				return res, err
			}
			res.Created++
			continue
		}
		var full models.Page
		if err := facades.Orm().Query().Where("id", cur.ID).First(&full); err != nil {
			return res, err
		}
		if full.Body == g.Body && full.Title == g.Title && full.Description == g.Description && full.Position == g.Position && full.Published {
			res.Unchanged++
			continue
		}
		if _, err := SavePage(p, in, &full); err != nil {
			return res, err
		}
		res.Updated++
	}
	for slug, pg := range bySlug {
		if !seen[slug] {
			if err := DeletePage(p, pg); err != nil {
				return res, err
			}
			res.Deleted++
		}
	}
	return res, nil
}
