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
