package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000080CreateAccessTables adds who may read a private project:
// members (users) and share links (signed-in-by-link readers).
type M20260924000080CreateAccessTables struct{}

func (r *M20260924000080CreateAccessTables) Signature() string {
	return "20260924000080_create_access_tables"
}

func (r *M20260924000080CreateAccessTables) Up() error {
	s := facades.Schema()
	if !s.HasTable("project_members") {
		if err := s.Create("project_members", func(t schema.Blueprint) {
			t.ID()
			t.UnsignedBigInteger("project_id")
			t.UnsignedBigInteger("user_id")
			t.TimestampsTz()
			t.Unique("project_id", "user_id")
			t.Index("user_id")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable("share_links") {
		if err := s.Create("share_links", func(t schema.Blueprint) {
			t.ID()
			t.UnsignedBigInteger("project_id")
			t.String("name")
			t.String("token_hash", 64)
			t.TimestampTz("expires_at").Nullable()
			t.TimestampTz("last_used_at").Nullable()
			t.UnsignedBigInteger("created_by").Default(0)
			t.TimestampsTz()
			t.Unique("token_hash")
			t.Index("project_id")
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260924000080CreateAccessTables) Down() error {
	if err := facades.Schema().DropIfExists("share_links"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("project_members")
}
