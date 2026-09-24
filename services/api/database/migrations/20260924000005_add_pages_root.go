package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000005AddPagesRoot marks folder index pages that make their
// folder a root: a sidebar tab with its own tree (fumadocs' root folders).
type M20260924000005AddPagesRoot struct{}

func (r *M20260924000005AddPagesRoot) Signature() string {
	return "20260924000005_add_pages_root"
}

func (r *M20260924000005AddPagesRoot) Up() error {
	if facades.Schema().HasColumn("pages", "root") {
		return nil
	}
	return facades.Schema().Table("pages", func(t schema.Blueprint) {
		t.Boolean("root").Default(false)
	})
}

func (r *M20260924000005AddPagesRoot) Down() error {
	return facades.Schema().DropColumns("pages", []string{"root"})
}
