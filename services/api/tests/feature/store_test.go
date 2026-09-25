package feature

import (
	"context"
	"fmt"
	"os"
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

func (s *StoreTestSuite) TestOpenAPIPages() {
	spec, err := os.ReadFile("../../app/openapi/testdata/petstore.yaml")
	s.Require().NoError(err)
	// A page from an earlier per-operation import, which the reference replaces.
	_, err = store.SavePage(s.project, store.PageInput{Slug: "api-reference/pets/listpets", Title: "Old", Body: "x"}, nil)
	s.Require().NoError(err)

	res, err := store.ImportOpenAPI(s.project, spec, "")
	s.Require().NoError(err)
	s.Equal(store.SyncResult{Created: 1, Deleted: 1}, res)

	view, err := store.ViewPage(s.project, "api-reference")
	s.Require().NoError(err)
	s.Equal("openapi", view.Kind)
	s.Equal("Swagger Petstore", view.Title)
	s.Contains(string(view.API), `"id":"tag/pets/POST/pets"`)
	s.Contains(view.Markdown, "### Create a pet")
	s.Empty(view.HTML)

	hits, err := store.Search(s.project, "Create a pet", 10)
	s.Require().NoError(err)
	var hashes []string
	for _, h := range hits {
		hashes = append(hashes, h.Hash)
	}
	s.Contains(hashes, "tag/pets/POST/pets", "operations are heading results")

	res, err = store.ImportOpenAPI(s.project, spec, "")
	s.Require().NoError(err)
	s.Equal(1, res.Unchanged)

	// Saving through the admin keeps the kind; a broken spec is rejected.
	pg, err := store.FindPage(s.project, "api-reference")
	s.Require().NoError(err)
	_, err = store.SavePage(s.project, store.PageInput{Slug: pg.Slug, Title: pg.Title, Body: "openapi: 2.0"}, &pg)
	s.Error(err)

	// Spec files sync as OpenAPI pages.
	sync, err := store.SyncPages(s.project, fstest.MapFS{"shop.openapi.yaml": {Data: spec}}, false)
	s.Require().NoError(err)
	s.Equal(1, sync.Created)
	shop, err := store.FindPage(s.project, "shop")
	s.Require().NoError(err)
	s.Equal(store.KindOpenAPI, shop.Kind)
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

func (s *StoreTestSuite) TestPrivateProjectAccess() {
	priv := false
	p, err := store.SaveProject(store.ProjectInput{Slug: s.project.Slug + "-private", Name: "Private", Public: &priv}, nil)
	s.Require().NoError(err)
	defer func() { s.NoError(store.DeleteProject(p)) }()

	_, err = store.ReadableProject(p.Slug, store.Reader{})
	s.ErrorIs(err, store.ErrNotFound, "anonymous readers do not see private projects")

	viewer := models.User{Name: "v", Email: p.Slug + "@example.com", Password: "x", Role: store.RoleViewer}
	s.Require().NoError(facades.Orm().Query().Create(&viewer))
	defer func() { _, _ = facades.Orm().Query().Delete(&viewer) }()

	_, err = store.ReadableProject(p.Slug, store.Reader{User: &viewer})
	s.ErrorIs(err, store.ErrNotFound, "non-members do not see it")
	domains, err := store.SetDomains(p, []string{"@Example.com"})
	s.Require().NoError(err)
	s.Equal([]string{"example.com"}, domains)
	got, err := store.ReadableProject(p.Slug, store.Reader{User: &viewer})
	s.Require().NoError(err, "a user at an allowed domain reads it")
	s.Equal(store.AccessDomain, got.Access)
	_, err = store.SetDomains(p, nil)
	s.Require().NoError(err)
	_, err = store.ReadableProject(p.Slug, store.Reader{User: &viewer})
	s.ErrorIs(err, store.ErrNotFound, "clearing the domains takes access away")
	s.Require().NoError(store.AddMember(p, viewer.ID))
	got, err = store.ReadableProject(p.Slug, store.Reader{User: &viewer})
	s.Require().NoError(err)
	s.Equal(store.AccessMember, got.Access)
	s.Require().NoError(store.RemoveMember(p, viewer.ID))

	token, link, err := store.CreateShare(p, "client", 7, viewer)
	s.Require().NoError(err)
	got, err = store.ReadableProject(p.Slug, store.Reader{Share: token})
	s.Require().NoError(err)
	s.Equal(store.AccessShare, got.Access)
	_, err = store.ReadableProject(s.project.Slug, store.Reader{Share: token})
	s.NoError(err, "the public project stays readable")
	s.Require().NoError(store.RevokeShare(p, link.ID))
	_, err = store.ReadableProject(p.Slug, store.Reader{Share: token})
	s.ErrorIs(err, store.ErrNotFound, "revoked links stop working")
}
