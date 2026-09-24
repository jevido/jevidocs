package store

import (
	"testing"

	"dev.jevido/jevidocs/services/api/app/models"
)

func TestLocales(t *testing.T) {
	p := models.Project{Locales: "en, NL,xx-!!,de", DefaultLocale: "en"}
	if got := Locales(p); len(got) != 3 || got[0] != "en" || got[1] != "nl" || got[2] != "de" {
		t.Errorf("Locales = %v", got)
	}
	for in, want := range map[string]string{"": "", "en": "", "nl": "nl", "NL": "nl", "fr": "", "api": ""} {
		if got := ResolveLocale(p, in); got != want {
			t.Errorf("ResolveLocale(%q) = %q, want %q", in, got, want)
		}
	}
	for file, want := range map[string][2]string{
		"guide.nl.md":        {"guide.md", "nl"},
		"guides/index.de.md": {"guides/index.md", "de"},
		"guide.md":           {"guide.md", ""},
		"v1.2.md":            {"v1.2.md", ""},
		"guide.en.md":        {"guide.md", ""},
	} {
		f, l := splitLocaleFile(p, file)
		if f != want[0] || l != want[1] {
			t.Errorf("splitLocaleFile(%q) = %q, %q; want %v", file, f, l, want)
		}
	}
	if got := normalizeLocales(" en,nl,nl, bad!"); got != "en,nl" {
		t.Errorf("normalizeLocales = %q", got)
	}
}
