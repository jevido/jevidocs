package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000070AddUserRoles adds users.role (existing users stay admins)
// and pages.updated_by (who last edited a page in the admin or over MCP).
type M20260924000070AddUserRoles struct{}

func (r *M20260924000070AddUserRoles) Signature() string {
	return "20260924000070_add_user_roles"
}

func (r *M20260924000070AddUserRoles) Up() error {
	s := facades.Schema()
	if !s.HasColumn("users", "role") {
		if err := s.Table("users", func(t schema.Blueprint) {
			t.String("role", 16).Default("admin")
		}); err != nil {
			return err
		}
	}
	if !s.HasColumn("pages", "updated_by") {
		return s.Table("pages", func(t schema.Blueprint) {
			t.UnsignedBigInteger("updated_by").Nullable()
		})
	}
	return nil
}

func (r *M20260924000070AddUserRoles) Down() error {
	if err := facades.Schema().DropColumns("users", []string{"role"}); err != nil {
		return err
	}
	return facades.Schema().DropColumns("pages", []string{"updated_by"})
}
