// Package database holds startup-time database chores.
package database

import (
	"fmt"

	"github.com/goravel/framework/database/migration"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// MigrateOnStart runs pending migrations when DB_MIGRATE_ON_START is true.
//
// It calls the migrator directly rather than `artisan migrate`, because that
// command prints a failure and still exits 0, which would let a container
// that could not migrate start serving. Here a failure is an error, and main
// exits on it.
func MigrateOnStart() error {
	if !facades.Config().GetBool("database.migrate_on_start") {
		return nil
	}
	migrator := migration.NewMigrator(
		facades.Artisan(),
		facades.Schema(),
		facades.Config().GetString("database.migrations.table"),
	)
	if err := migrator.Run(); err != nil {
		return fmt.Errorf("migrate on start: %w", err)
	}
	return nil
}
