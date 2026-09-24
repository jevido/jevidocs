package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// Ask answers a question from a project's docs (POST /api/projects/{project}/ask).
func (r *DocsController) Ask(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	if !store.AskEnabled() {
		return ctx.Response().Json(http.StatusNotFound, http.Json{"error": store.ErrAskDisabled.Error()})
	}
	var in struct {
		Question string `json:"question"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	res, err := store.Ask(ctx, p, in.Question, clientIP(ctx))
	if err != nil {
		var ve store.ValidationError
		if errors.As(err, &ve) || errors.Is(err, store.ErrTooManyAttempts) {
			return fail(ctx, err)
		}
		return ctx.Response().Json(http.StatusBadGateway, http.Json{"error": err.Error()})
	}
	return ok(ctx, res)
}
