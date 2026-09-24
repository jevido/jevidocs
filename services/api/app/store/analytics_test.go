package store

import (
	"strings"
	"testing"
)

func TestNormalizeQuery(t *testing.T) {
	for in, want := range map[string]string{
		"  Install   GUIDE ": "install guide",
		"a":                  "",
		"":                   "",
		"ab":                 "ab",
	} {
		if got := NormalizeQuery(in); got != want {
			t.Errorf("NormalizeQuery(%q) = %q, want %q", in, got, want)
		}
	}
	if got := NormalizeQuery(strings.Repeat("é", 150)); len([]rune(got)) != 100 {
		t.Errorf("not capped: %d runes", len([]rune(got)))
	}
}

func TestIsBot(t *testing.T) {
	if !IsBot("Googlebot/2.1") || !IsBot("") || !IsBot("Mozilla/5.0 HeadlessChrome") {
		t.Error("bot not detected")
	}
	if IsBot("Mozilla/5.0 (X11; Linux x86_64) Firefox/140.0") {
		t.Error("browser flagged as bot")
	}
}
