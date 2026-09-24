package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// Sitemap serves /api/projects/{project}/sitemap.xml.
func (r *DocsController) Sitemap(ctx http.Context) http.Response {
	p, err := publicProject(ctx)
	if err != nil {
		return fail(ctx, err)
	}
	s, err := store.Sitemap(p)
	if err != nil {
		return fail(ctx, err)
	}
	return text(ctx, "application/xml", s)
}
