package controllers

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/http/middleware"
	"dev.jevido/jevidocs/services/api/app/models"
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
	case errors.Is(err, store.ErrTooManyAttempts):
		return ctx.Response().Json(http.StatusTooManyRequests, http.Json{"error": err.Error()})
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

// reader is who is reading: an optional bearer token (signed-in user) and
// an optional share-link token in X-Jevidocs-Share.
func reader(ctx http.Context) store.Reader {
	return store.ReaderFrom(middleware.BearerToken(ctx), ctx.Request().Header("X-Jevidocs-Share", ""))
}

const privateKey = "jevidocs.private"

// publicProject loads the route's project if the reader may read it (public,
// or private with admin/member/share access), reading in the request's
// ?locale= (unknown or default locales read the default).
func publicProject(ctx http.Context) (models.Project, error) {
	p, err := store.ReadableProject(ctx.Request().Route("project"), reader(ctx))
	if err != nil {
		return p, err
	}
	if !p.Public {
		ctx.WithValue(privateKey, true)
	}
	return store.WithLocale(p, ctx.Request().Query("locale", "")), nil
}

// cacheControl keeps private projects out of shared caches.
func cacheControl(ctx http.Context) string {
	if private, _ := ctx.Value(privateKey).(bool); private {
		return "private, no-store"
	}
	return "public, max-age=60"
}
