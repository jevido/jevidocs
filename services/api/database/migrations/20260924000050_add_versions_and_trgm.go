package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000050AddVersionsAndTrgm groups projects into versions of one
// documentation site and enables pg_trgm for fuzzy search when the server
// has it. The extension is optional: the DO block swallows a missing or
// forbidden extension inside its own subtransaction, so the migration never
// fails on it and search simply skips the fuzzy step.
type M20260924000050AddVersionsAndTrgm struct{}

func (r *M20260924000050AddVersionsAndTrgm) Signature() string {
	return "20260924000050_add_versions_and_trgm"
}

func (r *M20260924000050AddVersionsAndTrgm) Up() error {
	s := facades.Schema()
	if !s.HasColumn("projects", "version_group") {
		if err := s.Table("projects", func(t schema.Blueprint) {
			t.String("version_group").Default("")
			t.String("version_label").Default("")
		}); err != nil {
			return err
		}
	}
	return s.Sql(`DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'pg_trgm') THEN
    BEGIN
      CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION WHEN OTHERS THEN
      RAISE NOTICE 'pg_trgm not enabled: %', SQLERRM;
    END;
  END IF;
END $$`)
}

func (r *M20260924000050AddVersionsAndTrgm) Down() error {
	return facades.Schema().DropColumns("projects", []string{"version_group", "version_label"})
}
