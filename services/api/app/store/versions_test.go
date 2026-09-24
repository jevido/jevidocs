package store

import "testing"

func TestCompareLabels(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v10", "v9", 1},
		{"1.10", "1.9", 1},
		{"v2", "v2", 0},
		{"v1", "v1.1", -1},
		{"beta", "alpha", 1},
	}
	for _, c := range cases {
		if got := compareLabels(c.a, c.b); got != c.want {
			t.Errorf("compareLabels(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestValidateVersion(t *testing.T) {
	if g, l, err := validateVersion(" My-Docs ", " v2 "); err != nil || g != "my-docs" || l != "v2" {
		t.Errorf("got %q %q %v", g, l, err)
	}
	if _, _, err := validateVersion("bad group!", ""); err == nil {
		t.Error("bad group accepted")
	}
	if _, _, err := validateVersion("", "<script>"); err == nil {
		t.Error("bad label accepted")
	}
}
