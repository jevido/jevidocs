package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000060AddLocales adds i18n: project locales and a per-page locale
// (” is the default locale). The page slug is then unique per locale. The
// new index is created before the old one is dropped, so the previous
// release (which only writes locale ”) keeps working during the rollout.
type M20260924000060AddLocales struct{}

func (r *M20260924000060AddLocales) Signature() string {
	return "20260924000060_add_locales"
}

func (r *M20260924000060AddLocales) Up() error {
	s := facades.Schema()
	if !s.HasColumn("projects", "locales") {
		if err := s.Table("projects", func(t schema.Blueprint) {
			t.String("locales").Default("")
			t.String("default_locale", 16).Default("en")
		}); err != nil {
			return err
		}
	}
	if !s.HasColumn("pages", "locale") {
		if err := s.Table("pages", func(t schema.Blueprint) {
			t.String("locale", 16).Default("")
		}); err != nil {
			return err
		}
	}
	for _, sql := range []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS pages_project_id_locale_slug_unique ON pages (project_id, locale, slug)`,
		`ALTER TABLE pages DROP CONSTRAINT IF EXISTS pages_project_id_slug_unique`,
		`DROP INDEX IF EXISTS pages_project_id_slug_unique`,
	} {
		if err := s.Sql(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260924000060AddLocales) Down() error {
	return nil
}
