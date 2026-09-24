package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/http/middleware"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/store"
)

// AdminController is the authenticated side used by apps/admin.
type AdminController struct{}

func NewAdminController() *AdminController { return &AdminController{} }

type userView struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func viewUser(u models.User) userView { return userView{u.ID, u.Name, u.Email} }

func (r *AdminController) Login(ctx http.Context) http.Response {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	token, u, err := store.Login(in.Email, in.Password)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"token": token, "user": viewUser(u)})
}

func (r *AdminController) Me(ctx http.Context) http.Response {
	return ok(ctx, http.Json{"user": viewUser(middleware.User(ctx))})
}

func (r *AdminController) Logout(ctx http.Context) http.Response {
	t := middleware.Token(ctx)
	if _, err := facades.Orm().Query().Delete(&t); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *AdminController) Stats(ctx http.Context) http.Response {
	projects, _ := facades.Orm().Query().Model(&models.Project{}).Count()
	pages, _ := facades.Orm().Query().Model(&models.Page{}).Count()
	tokens, _ := facades.Orm().Query().Model(&models.ApiToken{}).Where("kind", "api").Count()
	return ok(ctx, http.Json{"projects": projects, "pages": pages, "tokens": tokens})
}

func (r *AdminController) Projects(ctx http.Context) http.Response {
	ps, err := store.ListProjects(true)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, ps)
}

func (r *AdminController) Project(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, store.ViewProject(p))
}

func (r *AdminController) CreateProject(ctx http.Context) http.Response {
	var in store.ProjectInput
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	p, err := store.SaveProject(in, nil)
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, store.ViewProject(p))
}

func (r *AdminController) UpdateProject(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in store.ProjectInput
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	p, err = store.SaveProject(in, &p)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, store.ViewProject(p))
}

func (r *AdminController) DeleteProject(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	if err := store.DeleteProject(p); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

type adminPage struct {
	ID          uint   `json:"id"`
	Project     string `json:"project"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Position    int    `json:"position"`
	Section     string `json:"section"`
	Published   bool   `json:"published"`
	Body        string `json:"body,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func viewAdminPage(p models.Project, pg models.Page, withBody bool) adminPage {
	v := adminPage{ID: pg.ID, Project: p.Slug, Slug: pg.Slug, Title: pg.Title, Description: pg.Description,
		Icon: pg.Icon, Position: pg.Position, Section: pg.Section, Published: pg.Published}
	if withBody {
		v.Body = pg.Body
	}
	if pg.CreatedAt != nil {
		v.CreatedAt = pg.CreatedAt.ToIso8601String()
	}
	if pg.UpdatedAt != nil {
		v.UpdatedAt = pg.UpdatedAt.ToIso8601String()
	}
	return v
}

func (r *AdminController) Pages(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	pages, err := store.AdminPages(p)
	if err != nil {
		return fail(ctx, err)
	}
	out := make([]adminPage, 0, len(pages))
	for _, pg := range pages {
		out = append(out, viewAdminPage(p, pg, false))
	}
	return ok(ctx, out)
}

func (r *AdminController) ShowPage(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	pg, err := store.FindPageByID(p, routeID(ctx, "id"))
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, viewAdminPage(p, pg, true))
}

func (r *AdminController) CreatePage(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in store.PageInput
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	pg, err := store.SavePage(p, in, nil)
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, viewAdminPage(p, pg, true))
}

func (r *AdminController) UpdatePage(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	pg, err := store.FindPageByID(p, routeID(ctx, "id"))
	if err != nil {
		return fail(ctx, err)
	}
	var in store.PageInput
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	pg, err = store.SavePage(p, in, &pg)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, viewAdminPage(p, pg, true))
}

func (r *AdminController) DeletePage(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	pg, err := store.FindPageByID(p, routeID(ctx, "id"))
	if err != nil {
		return fail(ctx, err)
	}
	if err := store.DeletePage(p, pg); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}

func (r *AdminController) Preview(ctx http.Context) http.Response {
	var in struct {
		Body string `json:"body"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	out, err := store.Preview(in.Body)
	if err != nil {
		return fail(ctx, err)
	}
	toc := out.Toc
	if toc == nil {
		toc = []docs.TocItem{}
	}
	return ok(ctx, http.Json{"html": out.HTML, "toc": toc})
}

func (r *AdminController) Tokens(ctx http.Context) http.Response {
	ts, err := store.APITokens(middleware.User(ctx))
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, ts)
}

func (r *AdminController) CreateToken(ctx http.Context) http.Response {
	var in struct {
		Name string `json:"name"`
	}
	_ = ctx.Request().Bind(&in)
	if in.Name == "" {
		return fail(ctx, store.ValidationError{Msg: "name is required"})
	}
	plain, t, err := store.IssueToken(middleware.User(ctx), in.Name, "api")
	if err != nil {
		return fail(ctx, err)
	}
	return ctx.Response().Json(http.StatusCreated, http.Json{"id": t.ID, "name": t.Name, "token": plain})
}

func (r *AdminController) DeleteToken(ctx http.Context) http.Response {
	if err := store.RevokeToken(middleware.User(ctx), routeID(ctx, "id")); err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, http.Json{"ok": true})
}
