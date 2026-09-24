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
	feedback := controllers.NewFeedbackController()
	analytics := controllers.NewAnalyticsController()
	revisions := controllers.NewRevisionController()
	facades.Route().Get("/api/projects", docs.Projects)
	facades.Route().Prefix("/api/projects").Group(func(r route.Router) {
		r.Get("/{project}", docs.Project)
		r.Get("/{project}/page", docs.Page)
		r.Get("/{project}/page.md", docs.PageMarkdown)
		r.Get("/{project}/search", docs.Search)
		r.Get("/{project}/llms.txt", docs.LLMs)
		r.Get("/{project}/llms-full.txt", docs.LLMsFull)
		r.Get("/{project}/sitemap.xml", docs.Sitemap)
		r.Post("/{project}/ask", docs.Ask)
		r.Post("/{project}/feedback", feedback.Submit)
		r.Post("/{project}/views", analytics.View)
	})

	source := controllers.NewSourceController()
	facades.Route().Post("/api/hooks/github/{project}", source.Webhook)

	admin := controllers.NewAdminController()
	facades.Route().Post("/api/auth/login", admin.Login)
	assets := controllers.NewAssetController()
	users := controllers.NewUserController()
	facades.Route().Get("/api/assets/{id}/{name}", assets.Show)
	facades.Route().Middleware(middleware.Auth()).Group(func(r route.Router) {
		r.Get("/api/auth/me", admin.Me)
		r.Post("/api/auth/logout", admin.Logout)
		r.Put("/api/auth/password", users.Password)
		r.Prefix("/api/admin").Group(func(r route.Router) {
			r.Get("/stats", admin.Stats)
			r.Post("/preview", admin.Preview)
			r.Get("/projects", admin.Projects)
			r.Post("/projects", admin.CreateProject)
			r.Get("/projects/{project}", admin.Project)
			r.Put("/projects/{project}", admin.UpdateProject)
			r.Delete("/projects/{project}", admin.DeleteProject)
			r.Put("/projects/{project}/sync", admin.Sync)
			r.Get("/projects/{project}/source", source.Show)
			r.Put("/projects/{project}/source", source.Update)
			r.Post("/projects/{project}/source/sync", source.Sync)
			r.Post("/projects/{project}/openapi", controllers.NewOpenAPIController().Import)
			r.Put("/projects/{project}/order", admin.Order)
			r.Get("/projects/{project}/pages", admin.Pages)
			r.Post("/projects/{project}/pages", admin.CreatePage)
			r.Get("/projects/{project}/pages/{id}", admin.ShowPage)
			r.Put("/projects/{project}/pages/{id}", admin.UpdatePage)
			r.Delete("/projects/{project}/pages/{id}", admin.DeletePage)
			r.Get("/projects/{project}/pages/{id}/revisions", revisions.Index)
			r.Get("/projects/{project}/pages/{id}/revisions/{rid}", revisions.Show)
			r.Post("/projects/{project}/pages/{id}/revisions/{rid}/restore", revisions.Restore)
			r.Get("/projects/{project}/feedback", feedback.List)
			r.Get("/projects/{project}/insights", analytics.Insights)
			r.Get("/users", users.Index)
			r.Post("/users", users.Store)
			r.Delete("/users/{id}", users.Delete)
			r.Get("/projects/{project}/assets", assets.Index)
			r.Post("/projects/{project}/assets", assets.Upload)
			r.Delete("/projects/{project}/assets/{id}", assets.Delete)
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
