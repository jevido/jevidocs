package feature

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/suite"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/store"
	"dev.jevido/jevidocs/services/api/tests"
)

// StoreTestSuite exercises the store against a real Postgres. It needs a
// migrated database (task db:up + task api:dev once) and skips without one,
// which is the case inside the container build.
type StoreTestSuite struct {
	suite.Suite
	tests.TestCase
	project models.Project
}

func TestStoreTestSuite(t *testing.T) {
	if !databaseReady() {
		t.Skip("no migrated database; run `task db:up` and start the API once")
	}
	suite.Run(t, new(StoreTestSuite))
}

func databaseReady() (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	db, err := facades.Orm().DB()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if db.PingContext(ctx) != nil {
		return false
	}
	return facades.Schema().HasTable("pages")
}

func (s *StoreTestSuite) SetupTest() {
	pub := true
	p, err := store.SaveProject(store.ProjectInput{
		Slug: fmt.Sprintf("test-%d", time.Now().UnixNano()), Name: "Test", Public: &pub,
	}, nil)
	s.Require().NoError(err)
	s.project = p
}

func (s *StoreTestSuite) TearDownTest() {
	s.NoError(store.DeleteProject(s.project))
}

func (s *StoreTestSuite) TestPagesTreeSearchAndSync() {
	_, err := store.SavePage(s.project, store.PageInput{Slug: "", Title: "Home", Body: "Welcome to the **test** docs."}, nil)
	s.Require().NoError(err)
	_, err = store.SavePage(s.project, store.PageInput{Slug: "guides/install", Title: "Install",
		Body: "## Requirements\n\nYou need Postgres and a zebracorn.\n"}, nil)
	s.Require().NoError(err)

	_, err = store.SavePage(s.project, store.PageInput{Slug: "guides/install", Title: "Again"}, nil)
	s.Error(err, "duplicate slugs are rejected")

	tree, err := store.Tree(s.project)
	s.Require().NoError(err)
	s.Len(tree.Children, 2)
	s.Equal("folder", tree.Children[1].Type)

	view, err := store.ViewPage(s.project, "guides/install")
	s.Require().NoError(err)
	s.Equal("Install", view.Title)
	s.Contains(view.HTML, `id="requirements"`)
	s.Equal("", view.Previous.Slug)

	res, err := store.Search(s.project, "zebracorn", 10)
	s.Require().NoError(err)
	s.Require().NotEmpty(res)
	s.Equal("guides/install", res[0].Slug)

	sync, err := store.SyncPages(s.project, fstest.MapFS{
		"index.md":        {Data: []byte("---\ntitle: Home\n---\n\nWelcome to the **test** docs.")},
		"guides/new.md":   {Data: []byte("---\ntitle: New\nposition: 2\n---\n\nFresh.")},
		"guides/index.md": {Data: []byte("---\ntitle: Guides\n---\n\nAll guides.")},
	}, true)
	s.Require().NoError(err)
	s.Equal(2, sync.Created)
	s.Equal(1, sync.Deleted, "guides/install has no file and is pruned")

	_, err = store.FindPage(s.project, "guides/install")
	s.ErrorIs(err, store.ErrNotFound)
}

func (s *StoreTestSuite) TestTokens() {
	var u models.User
	s.Require().NoError(facades.Orm().Query().First(&u))
	if u.ID == 0 {
		s.T().Skip("no user to issue tokens for")
	}
	plain, tok, err := store.IssueToken(u, "test", "api")
	s.Require().NoError(err)
	got, _, err := store.UserForToken(plain)
	s.Require().NoError(err)
	s.Equal(u.ID, got.ID)
	s.NoError(store.RevokeToken(u, tok.ID))
	_, _, err = store.UserForToken(plain)
	s.ErrorIs(err, store.ErrNotFound)
}
