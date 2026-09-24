package store

import (
	"testing"

	"dev.jevido/jevidocs/services/api/app/models"
)

func TestRequiredRole(t *testing.T) {
	cases := []struct{ method, path, want string }{
		{"GET", "/api/admin/projects", RoleViewer},
		{"GET", "/api/admin/users", RoleViewer},
		{"POST", "/api/admin/preview", RoleViewer},
		{"POST", "/api/admin/tokens", RoleViewer},
		{"PUT", "/api/auth/password", RoleViewer},
		{"POST", "/api/admin/projects", RoleAdmin},
		{"PUT", "/api/admin/projects/demo", RoleAdmin},
		{"DELETE", "/api/admin/projects/demo", RoleAdmin},
		{"PUT", "/api/admin/projects/demo/source", RoleAdmin},
		{"POST", "/api/admin/users", RoleAdmin},
		{"DELETE", "/api/admin/users/3", RoleAdmin},
		{"POST", "/api/admin/projects/demo/pages", RoleEditor},
		{"PUT", "/api/admin/projects/demo/pages/4", RoleEditor},
		{"DELETE", "/api/admin/projects/demo/assets/1", RoleEditor},
		{"POST", "/api/admin/projects/demo/source/sync", RoleEditor},
		{"PUT", "/api/admin/projects/demo/order", RoleEditor},
		{"POST", "/api/admin/something-new", RoleEditor},
	}
	for _, c := range cases {
		if got := RequiredRole(c.method, c.path); got != c.want {
			t.Errorf("%s %s = %s, want %s", c.method, c.path, got, c.want)
		}
	}
}

func TestHasRole(t *testing.T) {
	viewer, editor, legacy := models.User{Role: RoleViewer}, models.User{Role: RoleEditor}, models.User{}
	if HasRole(viewer, RoleEditor) || !HasRole(viewer, RoleViewer) {
		t.Error("viewer rights wrong")
	}
	if !HasRole(editor, RoleEditor) || HasRole(editor, RoleAdmin) {
		t.Error("editor rights wrong")
	}
	if !HasRole(legacy, RoleAdmin) {
		t.Error("users without a role are admins")
	}
	if ValidRole("root") {
		t.Error("unknown role accepted")
	}
}
