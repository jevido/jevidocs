package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// FeedbackController takes readers' "Was this page helpful?" answers and
// shows them to admins.
type FeedbackController struct{}

func NewFeedbackController() *FeedbackController { return &FeedbackController{} }

func (r *FeedbackController) Submit(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		Slug    string `json:"slug"`
		Helpful bool   `json:"helpful"`
		Message string `json:"message"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	// Over the rate limit the answer is dropped, but the reader still sees
	// "thanks": nothing useful to tell a spammer.
	if _, err := store.SubmitFeedback(p, in.Slug, in.Helpful, in.Message, clientIP(ctx)); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *FeedbackController) List(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	entries, totals, err := store.ProjectFeedback(p)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"entries": entries, "totals": totals})
}
