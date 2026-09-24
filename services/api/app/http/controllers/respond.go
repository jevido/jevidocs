package controllers

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/store"
)

// fail maps store errors to status codes and never leaks internal detail.
func fail(ctx http.Context, err error) http.Response {
	var ve store.ValidationError
	switch {
	case errors.As(err, &ve):
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{"error": ve.Msg})
	case errors.Is(err, store.ErrNotFound):
		return ctx.Response().Json(http.StatusNotFound, http.Json{"error": "not found"})
	case errors.Is(err, store.ErrBadCredentials):
		return ctx.Response().Json(http.StatusUnauthorized, http.Json{"error": err.Error()})
	}
	facades.Log().Errorf("request %s: %v", ctx.Request().Path(), err)
	return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": "internal error"})
}

func ok(ctx http.Context, v any) http.Response {
	return ctx.Response().Json(http.StatusOK, v)
}

func routeID(ctx http.Context, key string) uint {
	n, _ := strconv.ParseUint(ctx.Request().Route(key), 10, 64)
	return uint(n)
}
