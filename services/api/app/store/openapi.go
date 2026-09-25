package store

import (
	"strings"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/openapi"
)

// ImportOpenAPI makes the page at prefix an OpenAPI page for spec: one page
// holding the whole reference, created or updated in place. Pages below
// prefix are deleted, since the reference owns that part of the tree (older
// imports generated a page per operation there). Pages outside prefix are
// never touched.
func ImportOpenAPI(p models.Project, spec []byte, prefix string) (SyncResult, error) {
	var res SyncResult
	prefix = docs.NormalizeSlug(prefix)
	if prefix == "" {
		prefix = openapi.DefaultPrefix
	}
	if !pageSlugRe.MatchString(prefix) {
		return res, ValidationError{"prefix must be a path of lowercase segments, like api-reference"}
	}
	existing, err := AdminPages(p)
	if err != nil {
		return res, err
	}
	var cur *models.Page
	for _, pg := range existing {
		switch {
		case pg.Slug == prefix:
			var full models.Page
			if err := facades.Orm().Query().Where("id", pg.ID).First(&full); err != nil {
				return res, err
			}
			cur = &full
		case strings.HasPrefix(pg.Slug, prefix+"/"):
			if err := DeletePage(p, pg); err != nil {
				return res, err
			}
			res.Deleted++
		}
	}

	body := plainText(string(spec))
	kind, pub := KindOpenAPI, true
	in := PageInput{Slug: prefix, Body: body, Kind: &kind, Published: &pub, Position: 100}
	if cur == nil {
		_, err = SavePage(p, in, nil)
		res.Created++
		return res, err
	}
	// Title, description and position set in the admin survive re-imports.
	in.Position = cur.Position
	if cur.Kind == KindOpenAPI {
		in.Title, in.Description = cur.Title, cur.Description
		if cur.Body == body && cur.Published {
			res.Unchanged++
			return res, nil
		}
	}
	_, err = SavePage(p, in, cur)
	res.Updated++
	return res, err
}
