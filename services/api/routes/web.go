package routes

import (
	"github.com/goravel/framework/contracts/route"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/http/controllers"
	"dev.jevido/jevidocs/services/api/app/http/middleware"
)

// Web registers the HTTP routes. This service is a JSON API only; every UI
// lives in apps/.
func Web() {
	healthController := controllers.NewHealthController()
	facades.Route().Get("/health", healthController.Show)

	docs := controllers.NewDocsController()
	facades.Route().Get("/api/projects", docs.Projects)
	facades.Route().Prefix("/api/projects").Group(func(r route.Router) {
		r.Get("/{project}", docs.Project)
		r.Get("/{project}/page", docs.Page)
		r.Get("/{project}/page.md", docs.PageMarkdown)
		r.Get("/{project}/search", docs.Search)
		r.Get("/{project}/llms.txt", docs.LLMs)
		r.Get("/{project}/llms-full.txt", docs.LLMsFull)
	})

	admin := controllers.NewAdminController()
	facades.Route().Post("/api/auth/login", admin.Login)
	facades.Route().Middleware(middleware.Auth()).Group(func(r route.Router) {
		r.Get("/api/auth/me", admin.Me)
		r.Post("/api/auth/logout", admin.Logout)
		r.Prefix("/api/admin").Group(func(r route.Router) {
			r.Get("/stats", admin.Stats)
			r.Post("/preview", admin.Preview)
			r.Get("/projects", admin.Projects)
			r.Post("/projects", admin.CreateProject)
			r.Get("/projects/{project}", admin.Project)
			r.Put("/projects/{project}", admin.UpdateProject)
			r.Delete("/projects/{project}", admin.DeleteProject)
			r.Put("/projects/{project}/sync", admin.Sync)
			r.Get("/projects/{project}/pages", admin.Pages)
			r.Post("/projects/{project}/pages", admin.CreatePage)
			r.Get("/projects/{project}/pages/{id}", admin.ShowPage)
			r.Put("/projects/{project}/pages/{id}", admin.UpdatePage)
			r.Delete("/projects/{project}/pages/{id}", admin.DeletePage)
			r.Get("/tokens", admin.Tokens)
			r.Post("/tokens", admin.CreateToken)
			r.Delete("/tokens/{id}", admin.DeleteToken)
		})
	})

	// Hosted MCP server (streamable HTTP). POST carries tool calls; GET and
	// DELETE are answered by the transport itself.
	mcpController := controllers.NewMcpController()
	facades.Route().Any("/mcp", mcpController.Serve)
}
