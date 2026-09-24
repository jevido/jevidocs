package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// Order saves new sibling positions after a drag-and-drop in the admin:
// PUT /api/admin/projects/{project}/order {items: [{id, position}]}.
func (r *AdminController) Order(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		Items []store.OrderItem `json:"items"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	if err := store.Reorder(p, in.Items); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}
