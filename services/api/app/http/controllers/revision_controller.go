package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/http/middleware"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/store"
)

// RevisionController serves a page's history in the admin.
type RevisionController struct{}

func NewRevisionController() *RevisionController { return &RevisionController{} }

func (r *RevisionController) page(ctx http.Context) (models.Project, models.Page, error) {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return p, models.Page{}, err
	}
	pg, err := store.FindPageByID(p, routeID(ctx, "id"))
	return p, pg, err
}

func (r *RevisionController) Index(ctx http.Context) http.Response {
	_, pg, err := r.page(ctx)
	if err != nil {
		return fail(ctx, err)
	}
	revs, err := store.Revisions(pg)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, revs)
}

func (r *RevisionController) Show(ctx http.Context) http.Response {
	_, pg, err := r.page(ctx)
	if err != nil {
		return fail(ctx, err)
	}
	rev, err := store.FindRevision(pg, routeID(ctx, "rid"))
	if err != nil {
		return fail(ctx, err)
	}
	v := http.Json{"id": rev.ID, "title": rev.Title, "description": rev.Description, "body": rev.Body, "size": len(rev.Body)}
	if rev.CreatedAt != nil {
		v["created_at"] = rev.CreatedAt.ToIso8601String()
	}
	return ok(ctx, v)
}

func (r *RevisionController) Restore(ctx http.Context) http.Response {
	p, pg, err := r.page(ctx)
	if err != nil {
		return fail(ctx, err)
	}
	rev, err := store.FindRevision(pg, routeID(ctx, "rid"))
	if err != nil {
		return fail(ctx, err)
	}
	pg, err = store.RestoreRevision(p, pg, rev)
	if err != nil {
		return fail(ctx, err)
	}
	editor := middleware.User(ctx).ID
	store.MarkEditedBy(pg.ID, editor)
	pg.UpdatedBy = &editor
	return ok(ctx, viewAdminPage(p, pg, true))
}
