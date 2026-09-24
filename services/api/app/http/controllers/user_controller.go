package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/http/middleware"
	"dev.jevido/jevidocs/services/api/app/store"
)

// UserController manages admin accounts.
type UserController struct{}

func NewUserController() *UserController { return &UserController{} }

func (r *UserController) Index(ctx http.Context) http.Response {
	us, err := store.ListUsers()
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, us)
}

func (r *UserController) Store(ctx http.Context) http.Response {
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	u, err := store.CreateUser(in.Name, in.Email, in.Password)
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, store.ViewUser(u))
}

func (r *UserController) Delete(ctx http.Context) http.Response {
	if err := store.DeleteUser(middleware.User(ctx), routeID(ctx, "id")); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *UserController) Password(ctx http.Context) http.Response {
	var in struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	if err := store.ChangePassword(middleware.User(ctx), in.Current, in.New); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}
