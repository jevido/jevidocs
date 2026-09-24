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
	err := store.SyncProject(store.SelfProject, "jevidocs",
		"Documentation framework on Go, Goravel, Svelte 5 and PostgreSQL, with an admin, a REST API and an MCP server.",
		"https://github.com/jevido/jevidocs",
		[]store.Link{{Text: "Docs", URL: "/docs"}, {Text: "Admin", URL: "https://admin.jevidocs.jevido.app"}},
		content.Docs())
	if err != nil {
		return fmt.Errorf("sync docs: %w", err)
	}
	return nil
}
