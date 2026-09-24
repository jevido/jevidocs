package store

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// VersionLink is one version of a documentation site in the reader's
// version switcher.
type VersionLink struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

var versionLabelRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._ -]{0,31}$`)

// validateVersion normalises a project's version group (slug-like) and
// label (short free text such as "v2" or "1.x").
func validateVersion(group, label string) (string, string, error) {
	group = strings.ToLower(strings.TrimSpace(group))
	label = strings.TrimSpace(label)
	if group != "" && (!slugRe.MatchString(group) || len(group) > 64) {
		return "", "", ValidationError{"version group must be lowercase letters, digits and dashes"}
	}
	if label != "" && !versionLabelRe.MatchString(label) {
		return "", "", ValidationError{"version label must be up to 32 letters, digits, dots, dashes or spaces"}
	}
	return group, label, nil
}

// Versions lists the public projects in p's version group, newest label
// first, including p itself. Nil when p has no group or is alone in it.
func Versions(p models.Project) ([]VersionLink, error) {
	if p.VersionGroup == "" {
		return nil, nil
	}
	var ps []models.Project
	if err := facades.Orm().Query().Where("version_group", p.VersionGroup).Where("public", true).Get(&ps); err != nil {
		return nil, err
	}
	if len(ps) < 2 {
		return nil, nil
	}
	sort.SliceStable(ps, func(i, j int) bool {
		return compareLabels(labelOf(ps[i]), labelOf(ps[j])) > 0
	})
	out := make([]VersionLink, 0, len(ps))
	for _, v := range ps {
		out = append(out, VersionLink{Slug: v.Slug, Name: v.Name, Label: labelOf(v), URL: PageURL(v, "")})
	}
	return out, nil
}

func labelOf(p models.Project) string {
	if p.VersionLabel != "" {
		return p.VersionLabel
	}
	return p.Slug
}

// compareLabels orders labels "naturally": runs of digits compare as
// numbers, so v10 > v9 and 1.10 > 1.9. Returns -1, 0 or 1.
func compareLabels(a, b string) int {
	ta, tb := labelTokens(a), labelTokens(b)
	for i := 0; i < len(ta) && i < len(tb); i++ {
		x, y := ta[i], tb[i]
		nx, errx := strconv.Atoi(x)
		ny, erry := strconv.Atoi(y)
		switch {
		case errx == nil && erry == nil:
			if nx != ny {
				if nx < ny {
					return -1
				}
				return 1
			}
		case x != y:
			if strings.ToLower(x) < strings.ToLower(y) {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(ta) < len(tb):
		return -1
	case len(ta) > len(tb):
		return 1
	}
	return 0
}

func labelTokens(s string) []string {
	var out []string
	var cur strings.Builder
	digit := false
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		d := unicode.IsDigit(r)
		if cur.Len() > 0 && d != digit {
			flush()
		}
		digit = d
		cur.WriteRune(r)
	}
	flush()
	return out
}
