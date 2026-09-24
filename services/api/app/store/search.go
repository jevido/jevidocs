package store

import (
	"encoding/json"
	"html"
	"strings"
	"unicode"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

type SearchResult struct {
	Type      string `json:"type"` // page | heading
	Slug      string `json:"slug"`
	Hash      string `json:"hash"`
	Title     string `json:"title"`
	PageTitle string `json:"page_title"`
	Snippet   string `json:"snippet"`
	URL       string `json:"url"`
	// Fuzzy marks a trigram "did you mean" match (see search_fuzzy.go).
	Fuzzy bool `json:"fuzzy,omitempty"`
}

// maxQueryLen bounds the text sent to ILIKE and trigram matching.
const maxQueryLen = 200

// Markers around matches from ts_headline. The snippet is plain text that
// may contain "<", so it is escaped in Go and only then gets <mark> tags.
const (
	markOpen  = "\x02"
	markClose = "\x03"
)

// terms splits a query into lowercase words of letters and digits.
func terms(q string) []string {
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(q), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len(out) < 8 {
			out = append(out, w)
		}
	}
	return out
}

// Search finds pages by full-text search (every word, prefix matched) and
// the headings inside them that mention a word.
func Search(p models.Project, q string, limit int) ([]SearchResult, error) {
	if r := []rune(q); len(r) > maxQueryLen {
		q = string(r[:maxQueryLen])
	}
	words := terms(q)
	if len(words) == 0 {
		return []SearchResult{}, nil
	}
	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = w + ":*"
	}
	tsq := strings.Join(parts, " & ")

	type row struct {
		Slug     string
		Title    string
		Sections string
		Snippet  string
	}
	var rows []row
	err := facades.Orm().Query().Raw(`
		SELECT slug, title, sections,
		  ts_headline('english', plain, query, 'StartSel=`+markOpen+`, StopSel=`+markClose+`, MaxWords=26, MinWords=10, MaxFragments=1') AS snippet
		FROM pages, to_tsquery('english', ?) AS query
		WHERE project_id = ? AND published AND `+localeFilter("pages")+`
		  AND to_tsvector('english', title || ' ' || description || ' ' || plain) @@ query
		ORDER BY ts_rank(to_tsvector('english', title || ' ' || description || ' ' || plain), query) DESC
		LIMIT 10`, tsq, p.ID, p.Locale, p.Locale).Scan(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		// Words Postgres' stemmer mangles (names, code): plain substring match.
		err = facades.Orm().Query().Raw(`
			SELECT slug, title, sections, left(plain, 180) AS snippet FROM pages
			WHERE project_id = ? AND published AND `+localeFilter("pages")+` AND (title ILIKE ? OR plain ILIKE ?)
			ORDER BY position LIMIT 10`, p.ID, p.Locale, p.Locale, "%"+likeEscape(q)+"%", "%"+likeEscape(q)+"%").Scan(&rows)
		if err != nil {
			return nil, err
		}
	}
	if len(rows) == 0 {
		return fuzzySearch(p, q, limit)
	}

	out := []SearchResult{}
	for _, r := range rows {
		out = append(out, SearchResult{Type: "page", Slug: r.Slug, Title: r.Title, PageTitle: r.Title,
			Snippet: markSnippet(r.Snippet, words), URL: PageURL(p, r.Slug)})
		var secs []docs.Section
		_ = json.Unmarshal([]byte(r.Sections), &secs)
		n := 0
		for _, s := range secs {
			if s.ID == "" || n >= 3 {
				continue
			}
			hay := strings.ToLower(s.Title + " " + s.Text)
			if !containsAny(hay, words) {
				continue
			}
			out = append(out, SearchResult{Type: "heading", Slug: r.Slug, Hash: s.ID, Title: s.Title, PageTitle: r.Title,
				Snippet: markSnippet(excerpt(s.Text, words), words), URL: PageURL(p, r.Slug) + "#" + s.ID})
			n++
		}
		if len(out) >= limit {
			break
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func containsAny(hay string, words []string) bool {
	for _, w := range words {
		if strings.Contains(hay, w) {
			return true
		}
	}
	return false
}

// excerpt cuts ~160 characters of text around the first matching word.
func excerpt(text string, words []string) string {
	lower := strings.ToLower(text)
	at := -1
	for _, w := range words {
		if i := strings.Index(lower, w); i >= 0 && (at < 0 || i < at) {
			at = i
		}
	}
	r := []rune(text)
	if at < 0 {
		at = 0
	} else {
		at = len([]rune(text[:at]))
	}
	start := max(0, at-60)
	end := min(len(r), start+160)
	s := string(r[start:end])
	if start > 0 {
		s = "…" + s
	}
	if end < len(r) {
		s += "…"
	}
	return s
}

// markSnippet escapes s and wraps matches (from ts_headline markers, or the
// query words) in <mark>.
func markSnippet(s string, words []string) string {
	if strings.Contains(s, markOpen) {
		s = html.EscapeString(s)
		return strings.NewReplacer(markOpen, "<mark>", markClose, "</mark>").Replace(s)
	}
	var b strings.Builder
	lower := strings.ToLower(s)
	i := 0
	for i < len(s) {
		hit := ""
		for _, w := range words {
			if strings.HasPrefix(lower[i:], w) && len(w) > len(hit) {
				hit = w
			}
		}
		if hit != "" {
			b.WriteString("<mark>" + html.EscapeString(s[i:i+len(hit)]) + "</mark>")
			i += len(hit)
			continue
		}
		b.WriteString(html.EscapeString(s[i : i+1]))
		i++
	}
	return b.String()
}

// likeEscape makes % and _ in user input match literally in ILIKE (whose
// default escape character is a backslash).
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}

// localeFilter limits a pages query (table or alias t) to what a reader in
// one locale sees: that locale's pages, plus default-locale pages without a
// translation. It takes two arguments, both the locale (” = default).
func localeFilter(t string) string {
	return "(" + t + ".locale = ? OR (" + t + ".locale = '' AND " + t + ".slug NOT IN (SELECT tr.slug FROM pages tr WHERE tr.project_id = " + t + ".project_id AND tr.locale = ?)))"
}
