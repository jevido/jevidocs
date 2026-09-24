package store

import "testing"

func TestCleanAssetName(t *testing.T) {
	for in, want := range map[string]string{
		"../../etc/passwd":  "passwd",
		"My Screenshot.png": "My-Screenshot.png",
		"ünï©ode.svg":       "node.svg",
		"...":               "file",
	} {
		if got := cleanAssetName(in); got != want {
			t.Errorf("cleanAssetName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAssetContentType(t *testing.T) {
	cases := []struct{ name, declared, want string }{
		{"a.png", "image/png", "image/png"},
		{"a.png", "application/octet-stream", "image/png"},
		{"a.svg", "", "image/svg+xml"},
		{"a.html", "text/html", ""},
		{"a.exe", "application/octet-stream", ""},
	}
	for _, c := range cases {
		if got := AssetContentType(c.name, c.declared); got != c.want {
			t.Errorf("AssetContentType(%q, %q) = %q, want %q", c.name, c.declared, got, c.want)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if ValidatePassword("short") == nil || ValidatePassword("longenough") != nil {
		t.Error("password policy")
	}
}
