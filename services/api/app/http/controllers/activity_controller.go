package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// ActivityController lists recent page edits for the admin dashboard.
type ActivityController struct{}

func NewActivityController() *ActivityController { return &ActivityController{} }

type activityRow struct {
	Project   string `json:"project"`
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Locale    string `json:"locale"`
	UpdatedAt string `json:"updated_at"`
}

// Index returns the 20 most recently updated pages across all projects.
func (r *ActivityController) Index(ctx http.Context) http.Response {
	rows := []activityRow{}
	err := facades.Orm().Query().Raw(`
		SELECT projects.slug AS project, pages.id, pages.slug, pages.title, pages.locale,
		  to_char(pages.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS updated_at
		FROM pages JOIN projects ON projects.id = pages.project_id
		ORDER BY pages.updated_at DESC NULLS LAST, pages.id DESC
		LIMIT 20`).Scan(&rows)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, rows)
}
