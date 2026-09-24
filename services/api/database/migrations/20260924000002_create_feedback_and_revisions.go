package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/app/facades"
)

type M20260924000002CreateFeedbackAndRevisions struct{}

func (r *M20260924000002CreateFeedbackAndRevisions) Signature() string {
	return "20260924000002_create_feedback_and_revisions"
}

func (r *M20260924000002CreateFeedbackAndRevisions) Up() error {
	s := facades.Schema()
	if !s.HasTable("feedback") {
		if err := s.Create("feedback", func(t schema.Blueprint) {
			t.ID()
			t.UnsignedBigInteger("project_id")
			t.String("slug")
			t.Boolean("helpful")
			t.Text("message").Default("")
			t.TimestampsTz()
			t.Index("project_id", "slug")
		}); err != nil {
			return err
		}
	}
	if !s.HasTable("page_revisions") {
		if err := s.Create("page_revisions", func(t schema.Blueprint) {
			t.ID()
			t.UnsignedBigInteger("page_id")
			t.UnsignedBigInteger("project_id")
			t.String("title")
			t.Text("description").Default("")
			t.Text("body").Default("")
			t.UnsignedBigInteger("user_id").Nullable()
			t.TimestampsTz()
			t.Index("page_id")
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260924000002CreateFeedbackAndRevisions) Down() error {
	for _, t := range []string{"page_revisions", "feedback"} {
		if err := facades.Schema().DropIfExists(t); err != nil {
			return err
		}
	}
	return nil
}
