package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

// ProjectMember lets a non-admin user read a private project.
type ProjectMember struct {
	orm.Model
	ProjectID uint `json:"project_id"`
	UserID    uint `json:"user_id"`
}

// ShareLink lets anyone holding its token read a private project. Only the
// token's SHA-256 is stored.
type ShareLink struct {
	orm.Model
	ProjectID  uint             `json:"-"`
	Name       string           `json:"name"`
	TokenHash  string           `json:"-"`
	ExpiresAt  *carbon.DateTime `json:"expires_at"`
	LastUsedAt *carbon.DateTime `json:"last_used_at"`
	CreatedBy  uint             `json:"-"`
}
