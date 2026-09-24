// Package store is the application layer over the database: projects, pages,
// search, llms.txt. REST controllers and MCP tools both call it, so a tool
// and its route cannot drift apart in validation.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// ErrNotFound is returned when a project or page does not exist (or is not
// visible to the caller).
var ErrNotFound = errors.New("not found")

// ValidationError is a problem with the caller's input.
type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

// SelfProject is the project holding jevidocs' own documentation.
const SelfProject = "jevidocs"

type Link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type ProjectView struct {
	Slug          string   `json:"slug"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	GithubURL     string   `json:"github_url"`
	Links         []Link   `json:"links"`
	Public        bool     `json:"public"`
	Managed       bool     `json:"managed"`
	EditURL       string   `json:"edit_url"`
	Banner        string   `json:"banner"`
	Accent        string   `json:"accent"`
	LogoURL       string   `json:"logo_url"`
	VersionGroup  string   `json:"version_group"`
	VersionLabel  string   `json:"version_label"`
	Locales       []string `json:"locales"`
	DefaultLocale string   `json:"default_locale"`
	Locale        string   `json:"locale"`
	// Access is how the reader may read it: public, admin, member or share.
	Access    string        `json:"access,omitempty"`
	UpdatedAt string        `json:"updated_at"`
	Tree      *docs.Tree    `json:"tree,omitempty"`
	Versions  []VersionLink `json:"versions,omitempty"`
	// Ask is true when Ask AI is enabled on this API.
	Ask bool `json:"ask"`
}

func ViewProject(p models.Project) ProjectView {
	links := []Link{}
	_ = json.Unmarshal([]byte(p.Links), &links)
	v := ProjectView{Slug: p.Slug, Name: p.Name, Description: p.Description, GithubURL: p.GithubURL,
		Links: links, Public: p.Public, Managed: p.Managed, EditURL: p.EditURL, Banner: p.Banner,
		Accent: p.Accent, LogoURL: p.LogoURL, VersionGroup: p.VersionGroup, VersionLabel: p.VersionLabel,
		Locales: Locales(p), DefaultLocale: DefaultLocale(p), Locale: p.Locale, Access: p.Access}
	if p.UpdatedAt != nil {
		v.UpdatedAt = p.UpdatedAt.ToIso8601String()
	}
	return v
}

// ListProjects returns projects, only public ones unless all is set.
func ListProjects(all bool) ([]ProjectView, error) {
	var ps []models.Project
	q := facades.Orm().Query().Order("name asc")
	if !all {
		q = q.Where("public", true)
	}
	if err := q.Get(&ps); err != nil {
		return nil, err
	}
	out := make([]ProjectView, 0, len(ps))
	for _, p := range ps {
		out = append(out, ViewProject(p))
	}
	return out, nil
}

// FindProject loads a project by slug. Private projects need all.
func FindProject(slug string, all bool) (models.Project, error) {
	var p models.Project
	if err := facades.Orm().Query().Where("slug", strings.ToLower(slug)).First(&p); err != nil {
		return p, err
	}
	if p.ID == 0 || (!p.Public && !all) {
		return p, ErrNotFound
	}
	return p, nil
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type ProjectInput struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	GithubURL    string `json:"github_url"`
	Links        []Link `json:"links"`
	Public       *bool  `json:"public"`
	EditURL      string `json:"edit_url"`
	Banner       string `json:"banner"`
	Accent       string `json:"accent"`
	LogoURL      string `json:"logo_url"`
	VersionGroup string `json:"version_group"`
	VersionLabel string `json:"version_label"`
	// Locales is a comma list ("en,nl"); nil keeps the current value.
	Locales       *string `json:"locales"`
	DefaultLocale string  `json:"default_locale"`
}

// SaveProject creates (existing == nil) or updates a project.
func SaveProject(in ProjectInput, existing *models.Project) (models.Project, error) {
	var p models.Project
	if existing != nil {
		p = *existing
	} else {
		p.Public = true
	}
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	if existing == nil || in.Slug != "" {
		if !slugRe.MatchString(in.Slug) || len(in.Slug) > 64 {
			return p, ValidationError{"slug must be lowercase letters, digits and dashes"}
		}
		if in.Slug != p.Slug {
			var other models.Project
			_ = facades.Orm().Query().Where("slug", in.Slug).First(&other)
			if other.ID != 0 {
				return p, ValidationError{"a project with that slug already exists"}
			}
		}
		p.Slug = in.Slug
	}
	if strings.TrimSpace(in.Name) == "" {
		return p, ValidationError{"name is required"}
	}
	p.Name = strings.TrimSpace(in.Name)
	p.Description = strings.TrimSpace(in.Description)
	p.GithubURL = strings.TrimSpace(in.GithubURL)
	p.EditURL = strings.TrimSpace(in.EditURL)
	p.Banner = strings.TrimSpace(in.Banner)
	accent, logo, err := validateTheme(in.Accent, in.LogoURL)
	if err != nil {
		return p, err
	}
	p.Accent, p.LogoURL = accent, logo
	group, label, err := validateVersion(in.VersionGroup, in.VersionLabel)
	if err != nil {
		return p, err
	}
	p.VersionGroup, p.VersionLabel = group, label
	if in.Locales != nil {
		p.Locales = normalizeLocales(*in.Locales)
	}
	if d := strings.ToLower(strings.TrimSpace(in.DefaultLocale)); d != "" {
		if !localeRe.MatchString(d) {
			return p, ValidationError{"default_locale must look like en or pt-br"}
		}
		p.DefaultLocale = d
	} else if p.DefaultLocale == "" {
		p.DefaultLocale = "en"
	}
	if in.Links == nil {
		in.Links = []Link{}
	}
	var links []Link
	for _, l := range in.Links {
		if strings.TrimSpace(l.Text) != "" && strings.TrimSpace(l.URL) != "" {
			links = append(links, Link{strings.TrimSpace(l.Text), strings.TrimSpace(l.URL)})
		}
	}
	if links == nil {
		links = []Link{}
	}
	b, _ := json.Marshal(links)
	p.Links = string(b)
	if in.Public != nil {
		p.Public = *in.Public
	}
	if existing == nil {
		err = facades.Orm().Query().Create(&p)
	} else {
		err = facades.Orm().Query().Save(&p)
	}
	return p, err
}

// projectTables hold rows keyed by project_id that go with the project.
var projectTables = []string{"page_revisions", "feedback", "page_views", "search_queries", "assets", "project_members", "share_links"}

// DeleteProject removes a project, its pages and everything attached to it
// (assets would otherwise stay publicly reachable by ID).
func DeleteProject(p models.Project) error {
	for _, t := range projectTables {
		if _, err := facades.Orm().Query().Exec(`DELETE FROM `+t+` WHERE project_id = ?`, p.ID); err != nil {
			return err
		}
	}
	if _, err := facades.Orm().Query().Where("project_id", p.ID).Delete(&models.Page{}); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Delete(&p)
	return err
}

var pageListColumns = []string{"id", "project_id", "slug", "title", "description", "icon", "position", "section", "published", "root", "locale", "updated_by", "created_at", "updated_at"}

// pagesOf lists p's pages in p.Locale. For readers (published) a missing
// translation falls back to the default-locale page, like fumadocs; for
// writers only pages of exactly that locale count.
func pagesOf(p models.Project, published bool) ([]models.Page, error) {
	var pages []models.Page
	q := facades.Orm().Query().Select(pageListColumns...).Where("project_id", p.ID)
	if published {
		q = q.Where("published", true)
		if p.Locale != "" {
			q = q.Where("locale IN ?", []string{"", p.Locale})
		} else {
			q = q.Where("locale", "")
		}
	} else {
		q = q.Where("locale", p.Locale)
	}
	if err := q.Order("slug asc").Get(&pages); err != nil {
		return nil, err
	}
	if !published || p.Locale == "" {
		return pages, nil
	}
	bySlug := map[string]int{}
	out := pages[:0:0]
	for _, pg := range pages {
		if i, ok := bySlug[pg.Slug]; ok {
			if pg.Locale == p.Locale {
				out[i] = pg
			}
			continue
		}
		bySlug[pg.Slug] = len(out)
		out = append(out, pg)
	}
	return out, nil
}

// AllLocalePages lists every page of every locale, for the admin.
func AllLocalePages(p models.Project) ([]models.Page, error) {
	var pages []models.Page
	err := facades.Orm().Query().Select(pageListColumns...).Where("project_id", p.ID).Order("slug asc").Order("locale asc").Get(&pages)
	return pages, err
}

// Tree builds the sidebar tree of a project's published pages.
func Tree(p models.Project) (docs.Tree, error) {
	pages, err := pagesOf(p, true)
	if err != nil {
		return docs.Tree{}, err
	}
	metas := make([]docs.PageMeta, 0, len(pages))
	for _, pg := range pages {
		metas = append(metas, docs.PageMeta{Slug: pg.Slug, Title: pg.Title, Icon: pg.Icon, Position: pg.Position,
			Section: pg.Section, Description: pg.Description, Root: pg.Root})
	}
	return docs.BuildTree(p.Name, metas), nil
}

type PageView struct {
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	HTML        string         `json:"html"`
	Toc         []docs.TocItem `json:"toc"`
	Breadcrumbs []docs.Crumb   `json:"breadcrumbs"`
	Previous    *docs.Link     `json:"previous"`
	Next        *docs.Link     `json:"next"`
	Markdown    string         `json:"markdown"`
	Locale      string         `json:"locale"`
	// Fallback is set when the page has no translation in the requested
	// locale and the default-locale page is shown instead.
	Fallback bool `json:"fallback"`
	// Draft is set when an unpublished page is shown through a preview link.
	Draft     bool   `json:"draft,omitempty"`
	URL       string `json:"url"`
	EditURL   string `json:"edit_url"`
	UpdatedAt string `json:"updated_at"`
}

// FindPage loads one published page by slug.
//
// In a non-default locale it falls back to the default-locale page.
func FindPage(p models.Project, slug string) (models.Page, error) {
	var pg models.Page
	slug = docs.NormalizeSlug(slug)
	if p.Locale != "" {
		err := facades.Orm().Query().Where("project_id", p.ID).Where("locale", p.Locale).Where("slug", slug).Where("published", true).First(&pg)
		if err != nil || pg.ID != 0 {
			return pg, err
		}
	}
	err := facades.Orm().Query().Where("project_id", p.ID).Where("locale", "").Where("slug", slug).Where("published", true).First(&pg)
	if err == nil && pg.ID == 0 {
		err = ErrNotFound
	}
	return pg, err
}

// ViewPage is a page with everything the reader needs around it.
func ViewPage(p models.Project, slug string) (PageView, error) {
	pg, err := FindPage(p, slug)
	if err != nil {
		return PageView{}, err
	}
	return viewOf(p, pg)
}

func viewOf(p models.Project, pg models.Page) (PageView, error) {
	tree, err := Tree(p)
	if err != nil {
		return PageView{}, err
	}
	v := PageView{Slug: pg.Slug, Title: pg.Title, Description: pg.Description, Icon: pg.Icon, HTML: pg.HTML,
		Markdown: pg.Body, URL: PageURL(p, pg.Slug), Toc: []docs.TocItem{},
		Locale: p.Locale, Fallback: p.Locale != "" && pg.Locale != p.Locale}
	_ = json.Unmarshal([]byte(pg.Toc), &v.Toc)
	v.Breadcrumbs = tree.Breadcrumbs(pg.Slug, pg.Title)
	v.Previous, v.Next = tree.Neighbours(pg.Slug)
	v.EditURL = EditURL(p, pg)
	if pg.UpdatedAt != nil {
		v.UpdatedAt = pg.UpdatedAt.ToIso8601String()
	}
	return v, nil
}

// SiteURL is where the docs reader is served.
func SiteURL() string {
	return strings.TrimRight(facades.Config().GetString("app.site_url", "https://jevidocs.jevido.app"), "/")
}

// PageURL is the public address of a page in the docs reader.
func PageURL(p models.Project, slug string) string {
	base := SiteURL() + "/p/" + p.Slug
	if p.Slug == SelfProject {
		base = SiteURL() + "/docs"
	}
	if p.Locale != "" {
		base += "/" + p.Locale
	}
	if slug == "" {
		return base
	}
	return base + "/" + slug
}

type PageInput struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Position    int    `json:"position"`
	Section     string `json:"section"`
	Published   *bool  `json:"published"`
	Body        string `json:"body"`
	// Root makes this folder index page's folder a sidebar tab. Nil keeps
	// the current value (or the front matter's).
	Root *bool `json:"root"`
	// Locale of a new page ('' or the default = default locale); nil means
	// the project's request locale. Ignored on updates.
	Locale *string `json:"locale"`
	// SourcePath is set by file syncs; empty keeps the current value.
	SourcePath string `json:"-"`
}

var pageSlugRe = regexp.MustCompile(`^(?:[a-z0-9][a-z0-9._-]*)(?:/[a-z0-9][a-z0-9._-]*)*$`)

// SavePage validates and renders a page, then creates (existing == nil) or
// updates it.
func SavePage(p models.Project, in PageInput, existing *models.Page) (models.Page, error) {
	var pg models.Page
	if existing != nil {
		pg = *existing
	} else {
		pg.ProjectID = p.ID
		pg.Published = true
		pg.Locale = p.Locale
		if in.Locale != nil {
			pg.Locale = ResolveLocale(p, *in.Locale)
		}
	}
	slug := docs.NormalizeSlug(in.Slug)
	if slug != "" && !pageSlugRe.MatchString(slug) {
		return pg, ValidationError{"slug must be a path of lowercase segments, like guides/install (empty for the index)"}
	}
	if existing == nil || slug != pg.Slug {
		var other models.Page
		_ = facades.Orm().Query().Where("project_id", p.ID).Where("locale", pg.Locale).Where("slug", slug).First(&other)
		if other.ID != 0 {
			return pg, ValidationError{fmt.Sprintf("a page with slug %q already exists", slug)}
		}
	}
	fm, body := docs.SplitFrontMatter(in.Body)
	title := firstNonEmpty(in.Title, fm.Title)
	if title == "" {
		return pg, ValidationError{"title is required"}
	}
	prev := pg // the stored state, for the revision history
	r, err := docs.Render(body)
	if err != nil {
		return pg, err
	}
	pg.Slug, pg.Title = slug, title
	pg.Description = firstNonEmpty(in.Description, fm.Description)
	pg.Icon = firstNonEmpty(in.Icon, fm.Icon)
	pg.Section = firstNonEmpty(in.Section, fm.Section)
	pg.Position = in.Position
	if in.Position == 0 && fm.HasPosition {
		pg.Position = fm.Position
	}
	if in.Published != nil {
		pg.Published = *in.Published
	}
	pg.Body = body
	pg.HTML = r.HTML
	pg.Plain = r.Plain
	toc, _ := json.Marshal(nonNil(r.Toc))
	pg.Toc = string(toc)
	secs, _ := json.Marshal(nonNil(r.Sections))
	pg.Sections = string(secs)
	pg.RenderVersion = docs.RenderVersion
	switch {
	case in.Root != nil:
		pg.Root = *in.Root
	case fm.Root:
		pg.Root = true
	case existing == nil:
		pg.Root = false
	}
	if in.SourcePath != "" {
		pg.SourcePath = in.SourcePath
	}
	if existing == nil {
		err = facades.Orm().Query().Create(&pg)
	} else {
		err = facades.Orm().Query().Save(&pg)
	}
	if err == nil {
		touch(p)
		if existing != nil && RevisionWorthy(prev, pg) {
			recordRevision(p, prev)
		}
	}
	return pg, err
}

func touch(p models.Project) {
	_, _ = facades.Orm().Query().Exec(`UPDATE projects SET updated_at = now() WHERE id = ?`, p.ID)
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
	}
	return ""
}

// AdminPages lists every page of a project, including unpublished ones.
func AdminPages(p models.Project) ([]models.Page, error) { return pagesOf(p, false) }

// FindPageByID loads any page of a project by ID.
func FindPageByID(p models.Project, id uint) (models.Page, error) {
	var pg models.Page
	err := facades.Orm().Query().Where("project_id", p.ID).Where("id", id).First(&pg)
	if err == nil && pg.ID == 0 {
		err = ErrNotFound
	}
	return pg, err
}

// DeletePage removes a page.
//
// Its revisions go with it. Feedback and views are keyed by slug and stay:
// a page recreated at the same slug keeps its history there.
func DeletePage(p models.Project, pg models.Page) error {
	if _, err := facades.Orm().Query().Exec(`DELETE FROM page_revisions WHERE page_id = ?`, pg.ID); err != nil {
		return err
	}
	_, err := facades.Orm().Query().Delete(&pg)
	if err == nil {
		touch(p)
	}
	return err
}

// Preview renders Markdown without saving it.
func Preview(body string) (docs.Rendered, error) {
	_, b := docs.SplitFrontMatter(body)
	return docs.Render(b)
}

// EditURL is the "Edit this page" link of pg, or "" when the project has no
// template.
func EditURL(p models.Project, pg models.Page) string {
	if p.EditURL == "" {
		return ""
	}
	path := pg.SourcePath
	if path == "" && p.Managed {
		// Pages of a file-synced project without a file were generated
		// (e.g. from OpenAPI); there is nothing to edit.
		return ""
	}
	if path == "" {
		path = pg.Slug + ".md"
		if pg.Slug == "" {
			path = "index.md"
		}
	}
	if !strings.Contains(p.EditURL, "{path}") {
		return strings.TrimRight(p.EditURL, "/") + "/" + path
	}
	return strings.ReplaceAll(p.EditURL, "{path}", path)
}
