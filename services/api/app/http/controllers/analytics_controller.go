package controllers

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// AnalyticsController counts page views and serves the Insights summary.
type AnalyticsController struct{}

func NewAnalyticsController() *AnalyticsController { return &AnalyticsController{} }

// View counts one page view. The site sends it with navigator.sendBeacon as
// text/plain (no CORS preflight), so the body is parsed as JSON whatever the
// content type says. Ignored views still answer ok.
func (r *AnalyticsController) View(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		Slug string `json:"slug"`
	}
	body, err := io.ReadAll(io.LimitReader(ctx.Request().Origin().Body, 4096))
	if err != nil || json.Unmarshal(body, &in) != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	if _, err := store.RecordView(p, in.Slug, clientIP(ctx), ctx.Request().Header("User-Agent", "")); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

// Insights summarises views, searches and feedback for the admin.
func (r *AnalyticsController) Insights(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	days, _ := strconv.Atoi(ctx.Request().Query("days", "30"))
	in, err := store.ProjectInsights(p, days)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, in)
}
