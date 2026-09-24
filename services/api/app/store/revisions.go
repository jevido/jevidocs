package store

import (
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// MaxRevisions is how many revisions are kept per page.
const MaxRevisions = 50

// RevisionWorthy reports whether saving next over prev changes content a
// reader sees, which is what the history keeps.
func RevisionWorthy(prev, next models.Page) bool {
	return prev.Body != next.Body || prev.Title != next.Title || prev.Description != next.Description
}

func recordRevision(p models.Project, prev models.Page) {
	rev := models.PageRevision{PageID: prev.ID, ProjectID: p.ID, Title: prev.Title, Description: prev.Description, Body: prev.Body}
	if err := facades.Orm().Query().Create(&rev); err != nil {
		facades.Log().Errorf("record revision of page %d: %v", prev.ID, err)
		return
	}
	_, _ = facades.Orm().Query().Exec(`DELETE FROM page_revisions WHERE page_id = ? AND id NOT IN
		(SELECT id FROM page_revisions WHERE page_id = ? ORDER BY id DESC LIMIT ?)`, prev.ID, prev.ID, MaxRevisions)
}

type RevisionSummary struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	Size      int    `json:"size"`
}

// Revisions lists a page's revisions, newest first.
func Revisions(pg models.Page) ([]RevisionSummary, error) {
	var revs []models.PageRevision
	if err := facades.Orm().Query().Where("page_id", pg.ID).Order("id desc").Get(&revs); err != nil {
		return nil, err
	}
	out := []RevisionSummary{}
	for _, r := range revs {
		s := RevisionSummary{ID: r.ID, Title: r.Title, Size: len(r.Body)}
		if r.CreatedAt != nil {
			s.CreatedAt = r.CreatedAt.ToIso8601String()
		}
		out = append(out, s)
	}
	return out, nil
}

// FindRevision loads one revision of pg.
func FindRevision(pg models.Page, id uint) (models.PageRevision, error) {
	var r models.PageRevision
	err := facades.Orm().Query().Where("page_id", pg.ID).Where("id", id).First(&r)
	if err == nil && r.ID == 0 {
		err = ErrNotFound
	}
	return r, err
}

// RestoreRevision saves the revision's content over the page; the state it
// replaces becomes a revision itself.
func RestoreRevision(p models.Project, pg models.Page, rev models.PageRevision) (models.Page, error) {
	pub := pg.Published
	return SavePage(p, PageInput{Slug: pg.Slug, Title: rev.Title, Description: rev.Description, Icon: pg.Icon,
		Position: pg.Position, Section: pg.Section, Published: &pub, Body: rev.Body}, &pg)
}
