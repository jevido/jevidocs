package models

import "github.com/goravel/framework/database/orm"

// Asset is an uploaded image or file of a project, deduplicated by SHA-256.
type Asset struct {
	orm.Model
	ProjectID   uint   `json:"-"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Sha256      string `gorm:"column:sha256" json:"sha256"`
	Data        []byte `json:"-"`
}
