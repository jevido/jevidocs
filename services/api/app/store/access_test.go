package store

import (
	"reflect"
	"testing"
)

func TestNormalizeDomains(t *testing.T) {
	got, err := NormalizeDomains([]string{"@Example.com", " example.org ", "example.com", ""})
	if err != nil || !reflect.DeepEqual(got, []string{"example.com", "example.org"}) {
		t.Fatalf("got %v, %v", got, err)
	}
	for _, bad := range []string{"example", "exa mple.com", "*.example.com", "user@example.com", "-x.example.com"} {
		if _, err := NormalizeDomains([]string{bad}); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestEmailInDomains(t *testing.T) {
	ds := []string{"example.com"}
	for email, want := range map[string]bool{
		"Ann@Example.COM":           true,
		"ann@mail.example.com":      false, // no subdomains
		"ann@example.com.evil.test": false,
		"ann@notexample.com":        false,
		"example.com@evil.test":     false,
		"ann":                       false,
	} {
		if got := emailInDomains(email, ds); got != want {
			t.Errorf("%s: got %v, want %v", email, got, want)
		}
	}
	if emailInDomains("ann@example.com", nil) {
		t.Error("no domains must match nothing")
	}
}
