package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000090AddProjectDomains lets a private project be read by every
// user whose email address is at one of its domains.
type M20260924000090AddProjectDomains struct{}

func (r *M20260924000090AddProjectDomains) Signature() string {
	return "20260924000090_add_project_domains"
}

func (r *M20260924000090AddProjectDomains) Up() error {
	if facades.Schema().HasColumn("projects", "allowed_domains") {
		return nil
	}
	return facades.Schema().Table("projects", func(t schema.Blueprint) {
		t.Text("allowed_domains").Default("")
	})
}

func (r *M20260924000090AddProjectDomains) Down() error {
	return facades.Schema().DropColumns("projects", []string{"allowed_domains"})
}
