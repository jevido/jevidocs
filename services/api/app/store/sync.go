package store

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// SyncProject makes a managed project's pages match the Markdown files in
// fsys: `guides/index.md` becomes slug `guides`, `index.md` the project
// index. Pages without a file are removed, so the files are the source of
// truth. Unchanged pages are not rewritten.
func SyncProject(in ProjectInput, fsys fs.FS) error {
	p, err := FindProject(in.Slug, true)
	if err != nil && err != ErrNotFound {
		return err
	}
	pub := true
	in.Public = &pub
	if p.ID == 0 {
		p, err = SaveProject(in, nil)
	} else {
		in.Slug = ""
		p, err = SaveProject(in, &p)
	}
	if err != nil {
		return err
	}
	if !p.Managed {
		p.Managed = true
		if err := facades.Orm().Query().Save(&p); err != nil {
			return err
		}
	}

	_, err = SyncPages(p, fsys, true)
	return err
}

// SyncResult counts what SyncPages changed.
type SyncResult struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
	Deleted   int `json:"deleted"`
}

// SyncPages makes p's pages match the Markdown files in fsys. With prune,
// pages that have no file are deleted.
func SyncPages(p models.Project, fsys fs.FS, prune bool) (SyncResult, error) {
	var res SyncResult
	existing, err := AdminPages(p)
	if err != nil {
		return res, err
	}
	bySlug := map[string]models.Page{}
	for _, pg := range existing {
		bySlug[pg.Slug] = pg
	}
	seen := map[string]bool{}

	err = fs.WalkDir(fsys, ".", func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || (path.Ext(file) != ".md" && path.Ext(file) != ".mdx") {
			return err
		}
		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		pageSlug := docs.NormalizeSlug(file)
		seen[pageSlug] = true
		fm, _ := docs.SplitFrontMatter(string(raw))
		title := fm.Title
		if title == "" {
			title = strings.TrimSuffix(path.Base(file), path.Ext(file))
		}
		in := PageInput{Slug: pageSlug, Title: title, Body: string(raw), SourcePath: file}
		if cur, ok := bySlug[pageSlug]; ok {
			var full models.Page
			if err := facades.Orm().Query().Where("id", cur.ID).First(&full); err != nil {
				return err
			}
			_, body := docs.SplitFrontMatter(string(raw))
			if full.Body == body && full.Title == title && full.Description == fm.Description &&
				full.Icon == fm.Icon && full.Section == fm.Section && full.Position == fm.Position && full.Published &&
				full.SourcePath == file {
				res.Unchanged++
				return nil
			}
			pub := true
			in.Published = &pub
			if _, err := SavePage(p, in, &full); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			res.Updated++
			return nil
		}
		if _, err = SavePage(p, in, nil); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		res.Created++
		return nil
	})
	if err != nil {
		return res, err
	}
	if !prune {
		return res, nil
	}
	for s, pg := range bySlug {
		if !seen[s] {
			if err := DeletePage(p, pg); err != nil {
				return res, err
			}
			res.Deleted++
		}
	}
	return res, nil
}
