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

// SyncProject makes a managed project's pages match the Markdown files (and
// OpenAPI documents named *.openapi.{json,yaml,yml}) in
// fsys: `guides/index.md` becomes slug `guides`, `index.md` the project
// index. Pages without a file are removed, so the files are the source of
// truth. Unchanged pages are not rewritten.
//
// Pages under keepPrefix (generated elsewhere, e.g. from OpenAPI) survive the
// pruning.
func SyncProject(in ProjectInput, fsys fs.FS, keepPrefix string) error {
	p, err := FindProject(in.Slug, true)
	if err != nil && err != ErrNotFound {
		return err
	}
	pub := true
	in.Public = &pub
	if p.ID == 0 {
		p, err = SaveProject(in, nil)
	} else {
		// Appearance is managed in the admin; the files only own the
		// project's identity, so keep what is already set.
		if in.Accent == "" {
			in.Accent = p.Accent
		}
		if in.LogoURL == "" {
			in.LogoURL = p.LogoURL
		}
		if in.Banner == "" {
			in.Banner = p.Banner
		}
		if in.VersionGroup == "" && in.VersionLabel == "" {
			in.VersionGroup, in.VersionLabel = p.VersionGroup, p.VersionLabel
		}
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

	_, err = syncPages(p, fsys, true, keepPrefix)
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
	return syncPages(p, fsys, prune, "")
}

func syncPages(p models.Project, fsys fs.FS, prune bool, keepPrefix string) (SyncResult, error) {
	var res SyncResult
	// Every locale at once: guide.nl.md is the nl page of slug "guide".
	existing, err := AllLocalePages(p)
	if err != nil {
		return res, err
	}
	key := func(locale, slug string) string { return locale + "\x00" + slug }
	bySlug := map[string]models.Page{}
	for _, pg := range existing {
		bySlug[key(pg.Locale, pg.Slug)] = pg
	}
	seen := map[string]bool{}

	err = fs.WalkDir(fsys, ".", func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		asMarkdown, isSpec := openAPIFile(file)
		if !isSpec && path.Ext(file) != ".md" && path.Ext(file) != ".mdx" {
			return nil
		}
		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		stripped, locale := splitLocaleFile(p, asMarkdown)
		if isSpec {
			return syncSpecFile(p, file, docs.NormalizeSlug(stripped), locale, plainText(string(raw)), bySlug[key(locale, docs.NormalizeSlug(stripped))], seen, key, &res)
		}
		pageSlug := docs.NormalizeSlug(stripped)
		seen[key(locale, pageSlug)] = true
		fm, _ := docs.SplitFrontMatter(string(raw))
		title := fm.Title
		if title == "" {
			title = strings.TrimSuffix(path.Base(file), path.Ext(file))
		}
		in := PageInput{Slug: pageSlug, Title: title, Body: string(raw), SourcePath: file, Locale: &locale}
		if cur, ok := bySlug[key(locale, pageSlug)]; ok {
			var full models.Page
			if err := facades.Orm().Query().Where("id", cur.ID).First(&full); err != nil {
				return err
			}
			_, body := docs.SplitFrontMatter(string(raw))
			if full.Body == body && full.Title == title && full.Description == fm.Description &&
				full.Icon == fm.Icon && full.Section == fm.Section && full.Position == fm.Position && full.Published &&
				full.SourcePath == file && full.Root == fm.Root {
				res.Unchanged++
				return nil
			}
			pub, root := true, fm.Root
			in.Published, in.Root = &pub, &root
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
	for k, pg := range bySlug {
		s := pg.Slug
		if keepPrefix != "" && (s == keepPrefix || strings.HasPrefix(s, keepPrefix+"/")) {
			continue
		}
		if !seen[k] {
			if err := DeletePage(p, pg); err != nil {
				return res, err
			}
			res.Deleted++
		}
	}
	return res, nil
}

// syncSpecFile creates or updates the OpenAPI page of one spec file. Its
// title, description and position come from the spec or the admin, as the
// file has no front matter.
func syncSpecFile(p models.Project, file, slug, locale, body string, cur models.Page, seen map[string]bool, key func(string, string) string, res *SyncResult) error {
	seen[key(locale, slug)] = true
	kind, pub := KindOpenAPI, true
	in := PageInput{Slug: slug, Body: body, SourcePath: file, Locale: &locale, Kind: &kind, Published: &pub}
	if cur.ID == 0 {
		if _, err := SavePage(p, in, nil); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		res.Created++
		return nil
	}
	var full models.Page
	if err := facades.Orm().Query().Where("id", cur.ID).First(&full); err != nil {
		return err
	}
	if full.Kind == KindOpenAPI && full.Body == body && full.Published && full.SourcePath == file {
		res.Unchanged++
		return nil
	}
	in.Position, in.Icon, in.Section = full.Position, full.Icon, full.Section
	if full.Kind == KindOpenAPI {
		in.Title, in.Description = full.Title, full.Description
	}
	if _, err := SavePage(p, in, &full); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	res.Updated++
	return nil
}
