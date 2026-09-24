package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000040AddProjectSource adds a project's GitHub source: the repo
// folder its pages sync from, the webhook secret and the last sync result.
type M20260924000040AddProjectSource struct{}

func (r *M20260924000040AddProjectSource) Signature() string {
	return "20260924000040_add_project_source"
}

func (r *M20260924000040AddProjectSource) Up() error {
	if facades.Schema().HasColumn("projects", "source_repo") {
		return nil
	}
	return facades.Schema().Table("projects", func(t schema.Blueprint) {
		t.String("source_repo").Default("")
		t.String("source_ref").Default("main")
		t.String("source_path").Default("docs")
		t.String("source_secret", 64).Default("")
		t.TimestampTz("source_synced_at").Nullable()
		t.Text("source_status").Default("")
	})
}

func (r *M20260924000040AddProjectSource) Down() error {
	return facades.Schema().DropColumns("projects", []string{"source_repo", "source_ref", "source_path", "source_secret", "source_synced_at", "source_status"})
}
