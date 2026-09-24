package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000004AddEditURLAndBanner adds a project's "edit this page" URL
// template and announcement banner, and each page's source file path.
type M20260924000004AddEditURLAndBanner struct{}

func (r *M20260924000004AddEditURLAndBanner) Signature() string {
	return "20260924000004_add_edit_url_and_banner"
}

func (r *M20260924000004AddEditURLAndBanner) Up() error {
	s := facades.Schema()
	if !s.HasColumn("projects", "edit_url") {
		if err := s.Table("projects", func(t schema.Blueprint) {
			t.String("edit_url").Default("")
			t.Text("banner").Default("")
		}); err != nil {
			return err
		}
	}
	if !s.HasColumn("pages", "source_path") {
		return s.Table("pages", func(t schema.Blueprint) {
			t.String("source_path").Default("")
		})
	}
	return nil
}

func (r *M20260924000004AddEditURLAndBanner) Down() error {
	if err := facades.Schema().DropColumns("projects", []string{"edit_url", "banner"}); err != nil {
		return err
	}
	return facades.Schema().DropColumns("pages", []string{"source_path"})
}
