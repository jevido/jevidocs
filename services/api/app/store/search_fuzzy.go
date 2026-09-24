package store

import (
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// fuzzyThreshold is the minimum trigram score for a "did you mean" hit.
const fuzzyThreshold = 0.3

// trgmAvailable reports whether pg_trgm is installed in this database. The
// migration enables it when the server allows; otherwise fuzzy search is
// simply skipped.
func trgmAvailable() bool {
	var n int64
	if err := facades.Orm().Query().Raw(`SELECT count(*) FROM pg_extension WHERE extname = 'pg_trgm'`).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

// fuzzySearch matches page titles and section headings by trigram
// similarity, for typos full-text search cannot find ("mermiad"). Results
// are marked Fuzzy so the reader can say "showing similar results".
func fuzzySearch(p models.Project, q string, limit int) ([]SearchResult, error) {
	if len([]rune(q)) < 3 || !trgmAvailable() {
		return []SearchResult{}, nil
	}
	type row struct {
		Slug      string
		PageTitle string
		Hash      string
		Title     string
		Snippet   string
		Score     float64
	}
	var rows []row
	err := facades.Orm().Query().Raw(`
		WITH candidates AS (
		  SELECT slug, title AS page_title, '' AS hash, title,
		         left(plain, 180) AS snippet,
		         greatest(similarity(lower(title), lower(?)), word_similarity(lower(?), lower(title))) AS score
		  FROM pages WHERE project_id = ? AND published
		  UNION ALL
		  SELECT pg.slug, pg.title, s->>'id', s->>'title',
		         left(s->>'text', 180),
		         greatest(similarity(lower(s->>'title'), lower(?)), word_similarity(lower(?), lower(s->>'title')))
		  FROM pages pg, jsonb_array_elements(CASE WHEN pg.sections LIKE '[%' THEN pg.sections::jsonb ELSE '[]'::jsonb END) s
		  WHERE pg.project_id = ? AND pg.published AND coalesce(s->>'id', '') <> ''
		)
		SELECT slug, page_title, hash, title, snippet, score FROM candidates
		WHERE score >= ? ORDER BY score DESC LIMIT ?`,
		q, q, p.ID, q, q, p.ID, fuzzyThreshold, limit).Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := []SearchResult{}
	for _, r := range rows {
		// Keep only hits close to the best one; weaker ones are noise.
		if r.Score < rows[0].Score*0.75 {
			break
		}
		res := SearchResult{Type: "page", Slug: r.Slug, Title: r.Title, PageTitle: r.PageTitle,
			Snippet: markSnippet(r.Snippet, nil), URL: PageURL(p, r.Slug), Fuzzy: true}
		if r.Hash != "" {
			res.Type, res.Hash, res.URL = "heading", r.Hash, res.URL+"#"+r.Hash
		}
		out = append(out, res)
	}
	return out, nil
}
