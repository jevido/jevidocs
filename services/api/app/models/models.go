// Package models holds the ORM models. Tables follow Goravel's plural snake
// case naming (projects, pages, users, api_tokens).
package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

// Project is one documentation site.
type Project struct {
	orm.Model
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GithubURL   string `gorm:"column:github_url" json:"github_url"`
	// Links is a JSON array of {text, url} for the docs navbar.
	Links  string `json:"-"`
	Public bool   `json:"public"`
	// EditURL is a template for "Edit this page" links; {path} is replaced
	// by the page's source path (or slug + ".md").
	EditURL string `gorm:"column:edit_url" json:"edit_url"`
	// Banner is an announcement shown above every page (Markdown inline).
	Banner string `json:"banner"`
	// Accent is a CSS colour or a preset name; LogoURL an https image.
	Accent  string `json:"accent"`
	LogoURL string `gorm:"column:logo_url" json:"logo_url"`
	// VersionGroup ties projects that are versions of one site together;
	// VersionLabel names this one ("v2").
	VersionGroup string `json:"version_group"`
	VersionLabel string `json:"version_label"`
	// GitHub source (see store.SyncFromGitHub). Never in public views.
	SourceRepo     string           `json:"-"`
	SourceRef      string           `json:"-"`
	SourcePath     string           `json:"-"`
	SourceSecret   string           `json:"-"`
	SourceSyncedAt *carbon.DateTime `json:"-"`
	SourceStatus   string           `json:"-"`
	// Managed projects are synced from files on start (the jevidocs docs).
	Managed bool `json:"managed"`
}

// Page is one Markdown page of a project. HTML, Toc, Sections and Plain are
// derived from Body whenever it is saved.
type Page struct {
	orm.Model
	ProjectID   uint   `json:"-"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Position    int    `json:"position"`
	Section     string `json:"section"`
	Published   bool   `json:"published"`
	Body        string `json:"body"`
	HTML        string `gorm:"column:html" json:"-"`
	Toc         string `json:"-"`
	Sections    string `json:"-"`
	Plain       string `json:"-"`
	// RenderVersion is docs.RenderVersion at the time HTML was produced.
	RenderVersion int `json:"-"`
	// SourcePath is the file a synced page came from, e.g. guides/index.md.
	SourcePath string `json:"source_path"`
	// Root makes a folder index page's folder a sidebar tab.
	Root bool `json:"root"`
}

// User is an admin account.
type User struct {
	orm.Model
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// ApiToken is a bearer token. Only its SHA-256 is stored. Kind is "session"
// (from logging in) or "api" (created for MCP and scripts).
type ApiToken struct {
	orm.Model
	UserID     uint             `json:"-"`
	Name       string           `json:"name"`
	Kind       string           `json:"kind"`
	TokenHash  string           `json:"-"`
	LastUsedAt *carbon.DateTime `json:"last_used_at"`
}
