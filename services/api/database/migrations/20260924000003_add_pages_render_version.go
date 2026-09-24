package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000003AddPagesRenderVersion records which renderer produced a
// page's stored HTML, so pages are re-rendered when the renderer changes.
type M20260924000003AddPagesRenderVersion struct{}

func (r *M20260924000003AddPagesRenderVersion) Signature() string {
	return "20260924000003_add_pages_render_version"
}

func (r *M20260924000003AddPagesRenderVersion) Up() error {
	if facades.Schema().HasColumn("pages", "render_version") {
		return nil
	}
	return facades.Schema().Table("pages", func(t schema.Blueprint) {
		t.Integer("render_version").Default(0)
	})
}

func (r *M20260924000003AddPagesRenderVersion) Down() error {
	return facades.Schema().DropColumns("pages", []string{"render_version"})
}
