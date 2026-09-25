package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260925000010AddPageKinds adds page kinds: "" is a Markdown page,
// "openapi" a page whose body is an OpenAPI document. For those, api holds
// the derived reference (JSON) and markdown the Markdown served to agents.
// All three default to empty, which is what the previous release expects.
type M20260925000010AddPageKinds struct{}

func (r *M20260925000010AddPageKinds) Signature() string {
	return "20260925000010_add_page_kinds"
}

func (r *M20260925000010AddPageKinds) Up() error {
	if facades.Schema().HasColumn("pages", "kind") {
		return nil
	}
	return facades.Schema().Table("pages", func(t schema.Blueprint) {
		t.String("kind", 20).Default("")
		t.Text("api").Default("")
		t.Text("markdown").Default("")
	})
}

func (r *M20260925000010AddPageKinds) Down() error {
	return facades.Schema().DropColumns("pages", []string{"kind", "api", "markdown"})
}
