package store

import (
	"strings"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// Roles, from least to most rights:
//   - viewer: reads everything in the admin, changes nothing
//   - editor: also edits pages, assets, imports and syncs
//   - admin: also manages projects' settings, and users
const (
	RoleViewer = "viewer"
	RoleEditor = "editor"
	RoleAdmin  = "admin"
)

var roleRank = map[string]int{RoleViewer: 1, RoleEditor: 2, RoleAdmin: 3}

// ValidRole reports whether r is a known role.
func ValidRole(r string) bool { return roleRank[r] > 0 }

// RoleOf is u's role; users from before roles existed are admins.
func RoleOf(u models.User) string {
	if ValidRole(u.Role) {
		return u.Role
	}
	return RoleAdmin
}

// HasRole reports whether u has at least role min.
func HasRole(u models.User, min string) bool {
	return roleRank[RoleOf(u)] >= roleRank[min]
}

// RequiredRole decides the minimum role for an authenticated request by
// method and path, so new admin routes default to "editor" for writes
// instead of silently being open to viewers.
func RequiredRole(method, path string) string {
	path = strings.TrimRight(path, "/")
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	method = strings.ToUpper(method)
	switch {
	case strings.HasPrefix(path, "/api/auth/"):
		return RoleViewer // me, logout, own password
	case strings.HasPrefix(path, "/api/admin/tokens"):
		return RoleViewer // own tokens; a token carries its owner's role
	case method == "GET" || method == "HEAD" || method == "OPTIONS":
		return RoleViewer
	case path == "/api/admin/preview":
		return RoleViewer // renders, stores nothing
	case strings.HasPrefix(path, "/api/admin/users"):
		return RoleAdmin
	case len(parts) == 3 && path == "/api/admin/projects" && method == "POST":
		return RoleAdmin // create project
	case len(parts) == 4 && parts[2] == "projects":
		return RoleAdmin // update or delete project settings
	case len(parts) == 5 && parts[2] == "projects" && parts[4] == "source" && method == "PUT":
		return RoleAdmin // repository and webhook secret
	case len(parts) >= 5 && parts[2] == "projects" && (parts[4] == "members" || parts[4] == "shares" || parts[4] == "domains"):
		return RoleAdmin // who may read a private project
	}
	return RoleEditor
}

// SetRole changes a user's role, keeping at least one admin.
func SetRole(id uint, role string) error {
	if !ValidRole(role) {
		return ValidationError{"role must be viewer, editor or admin"}
	}
	var u models.User
	if err := facades.Orm().Query().Where("id", id).First(&u); err != nil {
		return err
	}
	if u.ID == 0 {
		return ErrNotFound
	}
	if RoleOf(u) == RoleAdmin && role != RoleAdmin {
		if n, err := adminCount(); err != nil {
			return err
		} else if n <= 1 {
			return ValidationError{"cannot demote the last admin"}
		}
	}
	_, err := facades.Orm().Query().Exec(`UPDATE users SET role = ?, updated_at = now() WHERE id = ?`, role, id)
	return err
}

func adminCount() (int64, error) {
	return facades.Orm().Query().Model(&models.User{}).Where("role = ? OR role = '' OR role IS NULL", RoleAdmin).Count()
}

// MarkEditedBy records who last edited a page.
func MarkEditedBy(pageID, userID uint) {
	if pageID == 0 || userID == 0 {
		return
	}
	_, _ = facades.Orm().Query().Exec(`UPDATE pages SET updated_by = ? WHERE id = ?`, userID, pageID)
}

// UserNames maps user IDs to names, for "edited by".
func UserNames(ids []uint) map[uint]string {
	out := map[uint]string{}
	if len(ids) == 0 {
		return out
	}
	var us []models.User
	if err := facades.Orm().Query().WhereIn("id", toAny(ids)).Get(&us); err != nil {
		return out
	}
	for _, u := range us {
		out[u.ID] = u.Name
	}
	return out
}

func toAny(ids []uint) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}
