package migrations

import (
	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000010CreateAssetsTable stores uploaded images and files. The
// bytes live in Postgres (bytea) so backups and deploys need nothing else.
type M20260924000010CreateAssetsTable struct{}

func (r *M20260924000010CreateAssetsTable) Signature() string {
	return "20260924000010_create_assets_table"
}

func (r *M20260924000010CreateAssetsTable) Up() error {
	if facades.Schema().HasTable("assets") {
		return nil
	}
	return facades.Schema().Sql(`CREATE TABLE assets (
		id bigserial PRIMARY KEY,
		project_id bigint NOT NULL,
		name varchar(255) NOT NULL,
		content_type varchar(127) NOT NULL,
		size bigint NOT NULL,
		sha256 varchar(64) NOT NULL,
		data bytea NOT NULL,
		created_at timestamptz,
		updated_at timestamptz,
		UNIQUE (project_id, sha256)
	)`)
}

func (r *M20260924000010CreateAssetsTable) Down() error {
	return facades.Schema().DropIfExists("assets")
}
