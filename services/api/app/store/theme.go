package store

import (
	"regexp"
	"strings"
)

// ThemePresets are the named accents the site defines as data-preset.
var ThemePresets = map[string]bool{"neutral": true, "ocean": true, "purple": true, "emerald": true, "ruby": true}

// accentRe is deliberately strict: the accent is injected into CSS.
var accentRe = regexp.MustCompile(`^(#[0-9a-fA-F]{3}|#[0-9a-fA-F]{6}|hsl\(\s*\d{1,3}(\.\d+)?(deg)?\s+\d{1,3}(\.\d+)?%\s+\d{1,3}(\.\d+)?%\s*\)|oklch\(\s*\d(\.\d+)?%?\s+\d(\.\d+)?\s+\d{1,3}(\.\d+)?\s*\))$`)

// validateTheme normalises and checks a project's accent and logo URL.
func validateTheme(accent, logo string) (string, string, error) {
	accent = strings.TrimSpace(accent)
	logo = strings.TrimSpace(logo)
	if accent != "" && !ThemePresets[strings.ToLower(accent)] && !accentRe.MatchString(accent) {
		return "", "", ValidationError{"accent must be a preset (neutral, ocean, purple, emerald, ruby), #rgb, #rrggbb, hsl(h s% l%) or oklch(l c h)"}
	}
	if ThemePresets[strings.ToLower(accent)] {
		accent = strings.ToLower(accent)
	}
	if logo != "" && (!strings.HasPrefix(logo, "https://") || strings.ContainsAny(logo, " \"'<>")) {
		return "", "", ValidationError{"logo URL must start with https://"}
	}
	return accent, logo, nil
}
