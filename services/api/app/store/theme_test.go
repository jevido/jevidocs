package store

import "testing"

func TestValidateTheme(t *testing.T) {
	for _, ok := range []string{"", "#fff", "#12abEF", "hsl(262 83% 58%)", "oklch(0.6 0.2 250)", "Ocean"} {
		if _, _, err := validateTheme(ok, ""); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{"red", "#ffff", "hsl(1 2% 3%);color:red", "url(x)", "var(--x)"} {
		if _, _, err := validateTheme(bad, ""); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if _, _, err := validateTheme("", "http://x/logo.png"); err == nil {
		t.Error("http logo accepted")
	}
	if a, _, _ := validateTheme("Ocean", ""); a != "ocean" {
		t.Errorf("preset not normalised: %q", a)
	}
}
