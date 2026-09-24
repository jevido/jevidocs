package database

import (
	"fmt"

	"dev.jevido/jevidocs/services/api/app/store"
	"dev.jevido/jevidocs/services/api/content"
)

// PrepareOnStart creates the first admin (when configured and none exists)
// and syncs jevidocs' own documentation from the embedded content/ files.
func PrepareOnStart() error {
	if err := store.EnsureAdmin(); err != nil {
		return fmt.Errorf("ensure admin: %w", err)
	}
	err := store.SyncProject(store.ProjectInput{
		Slug:        store.SelfProject,
		Name:        "jevidocs",
		Description: "Documentation framework on Go, Goravel, Svelte 5 and PostgreSQL, with an admin, a REST API and an MCP server.",
		GithubURL:   "https://github.com/jevido/jevidocs",
		EditURL:     "https://github.com/jevido/jevidocs/blob/main/services/api/content/docs/{path}",
		Links:       []store.Link{{Text: "Docs", URL: "/docs"}, {Text: "Projects", URL: "/p"}, {Text: "Admin", URL: "https://admin.jevidocs.jevido.app"}},
	}, content.Docs())
	if err != nil {
		return fmt.Errorf("sync docs: %w", err)
	}
	if _, err := store.RerenderStale(); err != nil {
		return fmt.Errorf("re-render pages: %w", err)
	}
	return nil
}
