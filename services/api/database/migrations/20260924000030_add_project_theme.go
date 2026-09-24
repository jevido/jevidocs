package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000030AddProjectTheme adds a project's accent colour (or preset
// name) and logo URL for the docs reader.
type M20260924000030AddProjectTheme struct{}

func (r *M20260924000030AddProjectTheme) Signature() string {
	return "20260924000030_add_project_theme"
}

func (r *M20260924000030AddProjectTheme) Up() error {
	if facades.Schema().HasColumn("projects", "accent") {
		return nil
	}
	return facades.Schema().Table("projects", func(t schema.Blueprint) {
		t.String("accent").Default("")
		t.String("logo_url").Default("")
	})
}

func (r *M20260924000030AddProjectTheme) Down() error {
	return facades.Schema().DropColumns("projects", []string{"accent", "logo_url"})
}
