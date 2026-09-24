package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// DocsController serves the public, read-only side: projects, trees, pages,
// search and llms.txt.
type DocsController struct{}

func NewDocsController() *DocsController { return &DocsController{} }

func (r *DocsController) Projects(ctx http.Context) http.Response {
	ps, err := store.ListProjects(false)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, ps)
}

func (r *DocsController) Project(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	tree, err := store.Tree(p)
	if err != nil {
		return fail(ctx, err)
	}
	v := store.ViewProject(p)
	v.Tree = &tree
	return ok(ctx, v)
}

func (r *DocsController) Page(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	v, err := store.ViewPage(p, ctx.Request().Query("slug", ""))
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, v)
}

func (r *DocsController) Search(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	res, err := store.Search(p, ctx.Request().Query("q", ""), 20)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, res)
}

func text(ctx http.Context, contentType, body string) http.Response {
	return ctx.Response().Header("Content-Type", contentType+"; charset=utf-8").
		Header("Cache-Control", "public, max-age=60").
		String(http.StatusOK, body)
}

func (r *DocsController) LLMs(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	s, err := store.LLMsTxt(p)
	if err != nil {
		return fail(ctx, err)
	}
	return text(ctx, "text/plain", s)
}

func (r *DocsController) LLMsFull(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	s, err := store.LLMsFull(p)
	if err != nil {
		return fail(ctx, err)
	}
	return text(ctx, "text/plain", s)
}

func (r *DocsController) PageMarkdown(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), false)
	if err != nil {
		return fail(ctx, err)
	}
	pg, err := store.FindPage(p, ctx.Request().Query("slug", ""))
	if err != nil {
		return fail(ctx, err)
	}
	return text(ctx, "text/markdown", store.PageMarkdown(pg, store.PageURL(p, pg.Slug)))
}
