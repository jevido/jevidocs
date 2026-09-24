package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// ExportController serves project exports and draft preview links.
type ExportController struct{}

func NewExportController() *ExportController { return &ExportController{} }

// Export downloads the project as a zip of Markdown files.
func (r *ExportController) Export(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	data, err := store.Export(p)
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().
		Header("Content-Disposition", `attachment; filename="`+p.Slug+`.zip"`).
		Data(http.StatusOK, "application/zip", data)
}

// PreviewLink returns a 7-day link that shows the page even unpublished.
func (r *ExportController) PreviewLink(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	pg, err := store.FindPageByID(p, routeID(ctx, "id"))
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"url": store.PreviewLink(p, pg)})
}
