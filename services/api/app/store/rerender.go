package store

import (
	"encoding/json"

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
		r, err := docs.Render(pg.Body)
		if err != nil {
			return 0, err
		}
		toc, _ := json.Marshal(nonNil(r.Toc))
		secs, _ := json.Marshal(nonNil(r.Sections))
		if _, err := facades.Orm().Query().Exec(
			`UPDATE pages SET html = ?, toc = ?, sections = ?, plain = ?, render_version = ? WHERE id = ?`,
			r.HTML, string(toc), string(secs), r.Plain, docs.RenderVersion, pg.ID); err != nil {
			return 0, err
		}
	}
	return len(pages), nil
}
