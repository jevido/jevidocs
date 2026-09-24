package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/http/middleware"
	"dev.jevido/jevidocs/services/api/app/store"
)

// AccessController manages who may read a private project: members and
// share links.
type AccessController struct{}

func NewAccessController() *AccessController { return &AccessController{} }

func (r *AccessController) Members(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	ms, err := store.Members(p)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, ms)
}

func (r *AccessController) AddMember(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		UserID uint `json:"user_id"`
	}
	if err := ctx.Request().Bind(&in); err != nil || in.UserID == 0 {
		return fail(ctx, store.ValidationError{Msg: "user_id is required"})
	}
	if err := store.AddMember(p, in.UserID); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *AccessController) RemoveMember(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	if err := store.RemoveMember(p, routeID(ctx, "user")); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *AccessController) Shares(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	ss, err := store.Shares(p)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, ss)
}

func (r *AccessController) CreateShare(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		Name        string `json:"name"`
		ExpiresDays int    `json:"expires_days"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return fail(ctx, store.ValidationError{Msg: "invalid body"})
	}
	plain, link, err := store.CreateShare(p, in.Name, in.ExpiresDays, middleware.User(ctx))
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, http.Json{
		"id": link.ID, "name": link.Name, "token": plain, "url": store.ShareURL(p, plain),
	})
}

func (r *AccessController) RevokeShare(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	if err := store.RevokeShare(p, routeID(ctx, "id")); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *AccessController) Domains(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"domains": store.Domains(p)})
}

func (r *AccessController) SetDomains(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		Domains []string `json:"domains"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return fail(ctx, store.ValidationError{Msg: "invalid body"})
	}
	ds, err := store.SetDomains(p, in.Domains)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"domains": ds})
}
