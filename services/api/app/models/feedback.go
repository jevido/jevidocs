package models

import "github.com/goravel/framework/database/orm"

// Feedback is one "Was this page helpful?" answer from a reader.
type Feedback struct {
	orm.Model
	ProjectID uint   `json:"-"`
	Slug      string `json:"slug"`
	Helpful   bool   `json:"helpful"`
	Message   string `json:"message"`
}

// TableName keeps the table name uncountable ("feedback", not "feedbacks").
func (Feedback) TableName() string { return "feedback" }

// PageRevision is a page's state before a save changed it.
type PageRevision struct {
	orm.Model
	PageID      uint   `json:"-"`
	ProjectID   uint   `json:"-"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
	UserID      *uint  `json:"-"`
}
