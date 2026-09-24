package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

type M20260924000001CreateDocsTables struct{}

func (r *M20260924000001CreateDocsTables) Signature() string {
	return "20260924000001_create_docs_tables"
}

func (r *M20260924000001CreateDocsTables) Up() error {
	s := facades.Schema()
	if !s.HasTable("users") {
		if err := s.Create("users", func(t schema.Blueprint) {
			t.ID()
			t.String("name")
			t.String("email")
			t.String("password")
			t.TimestampsTz()
			t.Unique("email")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable("api_tokens") {
		if err := s.Create("api_tokens", func(t schema.Blueprint) {
			t.ID()
			t.UnsignedBigInteger("user_id")
			t.String("name")
			t.String("kind", 16)
			t.String("token_hash", 64)
			t.TimestampTz("last_used_at").Nullable()
			t.TimestampsTz()
			t.Unique("token_hash")
			t.Index("user_id")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable("projects") {
		if err := s.Create("projects", func(t schema.Blueprint) {
			t.ID()
			t.String("slug", 64)
			t.String("name")
			t.Text("description").Default("")
			t.String("github_url").Default("")
			t.Text("links").Default("[]")
			t.Boolean("public").Default(true)
			t.Boolean("managed").Default(false)
			t.TimestampsTz()
			t.Unique("slug")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable("pages") {
		if err := s.Create("pages", func(t schema.Blueprint) {
			t.ID()
			t.UnsignedBigInteger("project_id")
			t.String("slug")
			t.String("title")
			t.Text("description").Default("")
			t.String("icon").Default("")
			t.Integer("position").Default(0)
			t.String("section").Default("")
			t.Boolean("published").Default(true)
			t.Text("body").Default("")
			t.Text("html").Default("")
			t.Text("toc").Default("[]")
			t.Text("sections").Default("[]")
			t.Text("plain").Default("")
			t.TimestampsTz()
			t.Unique("project_id", "slug")
		}); err != nil {
			return err
		}
		// Full-text search over title, description and rendered text.
		if err := s.Sql(`CREATE INDEX IF NOT EXISTS pages_search_idx ON pages USING GIN (to_tsvector('english', title || ' ' || description || ' ' || plain))`); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260924000001CreateDocsTables) Down() error {
	for _, t := range []string{"pages", "projects", "api_tokens", "users"} {
		if err := facades.Schema().DropIfExists(t); err != nil {
			return err
		}
	}
	return nil
}
