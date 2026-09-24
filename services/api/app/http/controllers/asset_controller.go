package controllers

import (
	"errors"
	"io"
	nethttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// AssetController uploads, lists, deletes and serves project assets.
type AssetController struct{}

func NewAssetController() *AssetController { return &AssetController{} }

func (r *AssetController) Upload(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	req := ctx.Request().Origin()
	// One byte over the limit tells "too large" apart from "exactly 8 MB".
	req.Body = nethttp.MaxBytesReader(ctx.Response().Writer(), req.Body, store.MaxAssetSize+1<<20)
	file, header, err := req.FormFile("file")
	if err != nil {
		var tooBig *nethttp.MaxBytesError
		if errors.As(err, &tooBig) {
			return fail(ctx, store.ValidationError{Msg: "file is larger than 8 MB"})
		}
		return fail(ctx, store.ValidationError{Msg: "multipart field \"file\" is required"})
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, store.MaxAssetSize+1))
	if err != nil {
		return fail(ctx, store.ValidationError{Msg: "could not read upload"})
	}
	a, err := store.SaveAsset(p, header.Filename, header.Header.Get("Content-Type"), data)
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, store.ViewAsset(a))
}

func (r *AssetController) Index(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	as, err := store.ListAssets(p)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, as)
}

func (r *AssetController) Delete(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	if err := store.DeleteAsset(p, routeID(ctx, "id")); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

// Show serves an asset's bytes. The name in the URL is cosmetic; the ID
// decides. Assets are immutable (a new upload gets a new ID), so they cache
// forever.
func (r *AssetController) Show(ctx http.Context) http.Response {
	a, err := store.FindAsset(routeID(ctx, "id"))
	if err != nil {
		return fail(ctx, err)
	}
	if !store.AssetReadable(a, ctx.Request().Query("sig", "")) {
		return fail(ctx, store.ErrNotFound)
	}
	res := ctx.Response().
		Header("Cache-Control", "public, max-age=31536000, immutable").
		Header("X-Content-Type-Options", "nosniff").
		Header("Content-Disposition", "inline")
	if a.ContentType == "image/svg+xml" {
		// SVG can carry script; never let it run in our origin.
		res = res.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	}
	return res.Data(http.StatusOK, a.ContentType, a.Data)
}
