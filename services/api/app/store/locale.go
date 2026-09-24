package store

import (
	"regexp"
	"strings"

	"dev.jevido/jevidocs/services/api/app/models"
)

var localeRe = regexp.MustCompile(`^[a-z]{2,3}(?:-[a-z0-9]{2,8})?$`)

// Locales lists a project's languages, default first. A project without
// locales has exactly one, unnamed.
func Locales(p models.Project) []string {
	def := DefaultLocale(p)
	out := []string{def}
	for _, l := range strings.Split(p.Locales, ",") {
		l = strings.ToLower(strings.TrimSpace(l))
		if l != "" && l != def && localeRe.MatchString(l) {
			out = append(out, l)
		}
	}
	return out
}

// DefaultLocale is the language of pages stored with locale ”.
func DefaultLocale(p models.Project) string {
	if d := strings.ToLower(strings.TrimSpace(p.DefaultLocale)); d != "" {
		return d
	}
	return "en"
}

// ResolveLocale maps a requested locale to what pages store: ” for the
// default (or anything unknown), else the locale itself.
func ResolveLocale(p models.Project, l string) string {
	l = strings.ToLower(strings.TrimSpace(l))
	if l == "" || l == DefaultLocale(p) {
		return ""
	}
	for _, known := range Locales(p)[1:] {
		if known == l {
			return l
		}
	}
	return ""
}

// WithLocale returns p reading in locale l (see models.Project.Locale).
func WithLocale(p models.Project, l string) models.Project {
	p.Locale = ResolveLocale(p, l)
	return p
}

func normalizeLocales(s string) string {
	var out []string
	seen := map[string]bool{}
	for _, l := range strings.Split(s, ",") {
		l = strings.ToLower(strings.TrimSpace(l))
		if l != "" && !seen[l] && localeRe.MatchString(l) {
			seen[l] = true
			out = append(out, l)
		}
	}
	return strings.Join(out, ",")
}

// splitLocaleFile reads the fumadocs "dot" convention: guide.nl.md is the
// nl translation of guide.md. It returns the path without the locale and
// the locale (” when the file is in the default locale).
func splitLocaleFile(p models.Project, file string) (string, string) {
	ext := ""
	for _, e := range []string{".mdx", ".md"} {
		if strings.HasSuffix(file, e) {
			ext = e
		}
	}
	base := strings.TrimSuffix(file, ext)
	i := strings.LastIndex(base, ".")
	if i < 0 || strings.Contains(base[i:], "/") {
		return file, ""
	}
	if suffix := strings.ToLower(base[i+1:]); p.Locales != "" && suffix == DefaultLocale(p) {
		return base[:i] + ext, ""
	}
	if l := ResolveLocale(p, base[i+1:]); l != "" {
		return base[:i] + ext, l
	}
	return file, ""
}
