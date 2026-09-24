package controllers

import (
	"encoding/json"
	"io"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// SourceController connects a project to a folder in a GitHub repository:
// admin settings, "sync now", and the push webhook.
type SourceController struct{}

func NewSourceController() *SourceController { return &SourceController{} }

func (r *SourceController) Show(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, store.ViewSource(p))
}

func (r *SourceController) Update(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in store.SourceInput
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	p, err = store.SaveSource(p, in)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, store.ViewSource(p))
}

// Sync starts a sync and answers 202; poll Show for status and synced_at.
// Downloads can take longer than the request timeout.
func (r *SourceController) Sync(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	if p.SourceRepo == "" {
		return fail(ctx, store.ValidationError{Msg: "set a GitHub repository first"})
	}
	store.SyncFromGitHubAsync(p)
	return ctx.Response().Json(http.StatusAccepted, http.Json{"started": true, "source": store.ViewSource(p)})
}

// Webhook receives GitHub push events signed with the project's secret.
func (r *SourceController) Webhook(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil || p.SourceRepo == "" {
		return ctx.Response().Json(http.StatusNotFound, http.Json{"error": "not found"})
	}
	body, err := io.ReadAll(io.LimitReader(ctx.Request().Origin().Body, 5<<20))
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "unreadable body"})
	}
	if !store.VerifyGitHubSignature(p.SourceSecret, body, ctx.Request().Header("X-Hub-Signature-256", "")) {
		return ctx.Response().Json(http.StatusUnauthorized, http.Json{"error": "bad signature"})
	}
	switch ctx.Request().Header("X-GitHub-Event", "") {
	case "ping":
		return ok(ctx, http.Json{"pong": true})
	case "push":
	default:
		return ok(ctx, http.Json{"ignored": "not a push"})
	}
	var push struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &push); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid JSON"})
	}
	if push.Ref != "refs/heads/"+p.SourceRef {
		return ok(ctx, http.Json{"ignored": "push to another ref"})
	}
	store.SyncFromGitHubAsync(p)
	return ctx.Response().Json(http.StatusAccepted, http.Json{"started": true})
}
