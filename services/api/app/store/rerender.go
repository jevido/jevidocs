package store

import (
	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// RerenderStale re-renders every page whose stored HTML came from an older
// renderer (see docs.RenderVersion). Only derived columns change, so it is
// safe to run while the previous release is still serving.
func RerenderStale() (int, error) {
	var pages []models.Page
	if err := facades.Orm().Query().Where("render_version < ?", docs.RenderVersion).Get(&pages); err != nil {
		return 0, err
	}
	for _, pg := range pages {
		r, err := renderBody(pg.Kind, pg.Body)
		if err != nil {
			return 0, err
		}
		r.apply(&pg)
		if _, err := facades.Orm().Query().Exec(
			`UPDATE pages SET html = ?, toc = ?, sections = ?, plain = ?, api = ?, markdown = ?, render_version = ? WHERE id = ?`,
			pg.HTML, pg.Toc, pg.Sections, pg.Plain, pg.API, pg.Markdown, pg.RenderVersion, pg.ID); err != nil {
			return 0, err
		}
	}
	return len(pages), nil
}
