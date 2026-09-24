package store

import (
	"strings"
	"time"
	"unicode/utf8"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// viewLimiter caps page views counted per client address.
var viewLimiter = &RateLimiter{Limit: 300, Window: time.Hour}

// IsBot reports whether a User-Agent looks like a crawler.
func IsBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	if ua == "" {
		return true
	}
	for _, s := range []string{"bot", "crawler", "spider", "slurp", "headless", "preview"} {
		if strings.Contains(ua, s) {
			return true
		}
	}
	return false
}

// RecordView counts one view of a published page for today. Bots, unknown
// pages and addresses over the limit are ignored (counted reports false).
func RecordView(p models.Project, slug, ip, userAgent string) (counted bool, err error) {
	if IsBot(userAgent) {
		return false, nil
	}
	pg, err := FindPage(p, slug)
	if err != nil {
		return false, nil
	}
	if !viewLimiter.Allow(ip, time.Now()) {
		return false, nil
	}
	_, err = facades.Orm().Query().Exec(`INSERT INTO page_views (project_id, slug, day, views)
		VALUES (?, ?, CURRENT_DATE, 1)
		ON CONFLICT (project_id, slug, day) DO UPDATE SET views = page_views.views + 1`, p.ID, pg.Slug)
	return err == nil, err
}

// NormalizeQuery lowercases, collapses spaces and caps a search query at 100
// characters; "" means "do not record".
func NormalizeQuery(q string) string {
	q = strings.Join(strings.Fields(strings.ToLower(q)), " ")
	for utf8.RuneCountInString(q) > 100 {
		r := []rune(q)
		q = string(r[:100])
	}
	if utf8.RuneCountInString(q) < 2 {
		return ""
	}
	return q
}

// RecordSearch counts a public search and remembers its last result count.
// Runs in the background so search latency does not change.
func RecordSearch(p models.Project, q string, results int) {
	q = NormalizeQuery(q)
	if q == "" {
		return
	}
	go func() {
		defer func() { _ = recover() }()
		_, err := facades.Orm().Query().Exec(`INSERT INTO search_queries (project_id, query, day, count, results)
			VALUES (?, ?, CURRENT_DATE, 1, ?)
			ON CONFLICT (project_id, query, day) DO UPDATE SET count = search_queries.count + 1, results = EXCLUDED.results`,
			p.ID, q, results)
		if err != nil {
			facades.Log().Warningf("record search: %v", err)
		}
	}()
}

type DayViews struct {
	Day   string `json:"day"`
	Views int    `json:"views"`
}

type PageViews struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Views int    `json:"views"`
}

type SearchStat struct {
	Query   string `json:"query"`
	Count   int    `json:"count"`
	Results int    `json:"results"`
}

type Insights struct {
	Days               int          `json:"days"`
	ViewsTotal         int          `json:"views_total"`
	ViewsByDay         []DayViews   `json:"views_by_day"`
	TopPages           []PageViews  `json:"top_pages"`
	TopSearches        []SearchStat `json:"top_searches"`
	ZeroResultSearches []SearchStat `json:"zero_result_searches"`
	Feedback           struct {
		Helpful    int `json:"helpful"`
		NotHelpful int `json:"not_helpful"`
	} `json:"feedback"`
}

// ProjectInsights summarises the last days (1..365) of views, searches and
// feedback. views_by_day has one entry per day, zeros included.
func ProjectInsights(p models.Project, days int) (Insights, error) {
	if days < 1 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	in := Insights{Days: days, ViewsByDay: []DayViews{}, TopPages: []PageViews{}, TopSearches: []SearchStat{}, ZeroResultSearches: []SearchStat{}}
	q := facades.Orm().Query()
	since := days - 1

	var byDay []DayViews
	if err := q.Raw(`SELECT to_char(d::date, 'YYYY-MM-DD') AS day, COALESCE(sum(v.views), 0)::int AS views
		FROM generate_series(CURRENT_DATE - ?::int, CURRENT_DATE, interval '1 day') AS d
		LEFT JOIN page_views v ON v.day = d::date AND v.project_id = ?
		GROUP BY d ORDER BY d`, since, p.ID).Scan(&byDay); err != nil {
		return in, err
	}
	if byDay != nil {
		in.ViewsByDay = byDay
	}
	for _, d := range in.ViewsByDay {
		in.ViewsTotal += d.Views
	}

	var top []PageViews
	if err := q.Raw(`SELECT v.slug, COALESCE(max(pg.title), v.slug) AS title, sum(v.views)::int AS views
		FROM page_views v LEFT JOIN pages pg ON pg.project_id = v.project_id AND pg.slug = v.slug
		WHERE v.project_id = ? AND v.day > CURRENT_DATE - ?::int
		GROUP BY v.slug ORDER BY views DESC, v.slug LIMIT 10`, p.ID, days).Scan(&top); err != nil {
		return in, err
	}
	if top != nil {
		in.TopPages = top
	}

	var searches []SearchStat
	if err := q.Raw(`SELECT query, sum(count)::int AS count,
		(array_agg(results ORDER BY day DESC))[1] AS results
		FROM search_queries WHERE project_id = ? AND day > CURRENT_DATE - ?::int
		GROUP BY query ORDER BY count DESC, query LIMIT 50`, p.ID, days).Scan(&searches); err != nil {
		return in, err
	}
	for _, s := range searches {
		if s.Results == 0 {
			if len(in.ZeroResultSearches) < 10 {
				in.ZeroResultSearches = append(in.ZeroResultSearches, s)
			}
		} else if len(in.TopSearches) < 10 {
			in.TopSearches = append(in.TopSearches, s)
		}
	}

	var fb []struct {
		Helpful    int
		NotHelpful int
	}
	if err := q.Raw(`SELECT count(*) FILTER (WHERE helpful)::int AS helpful,
		count(*) FILTER (WHERE NOT helpful)::int AS not_helpful
		FROM feedback WHERE project_id = ? AND created_at > now() - make_interval(days => ?)`, p.ID, days).Scan(&fb); err != nil {
		return in, err
	}
	if len(fb) > 0 {
		in.Feedback.Helpful, in.Feedback.NotHelpful = fb[0].Helpful, fb[0].NotHelpful
	}
	return in, nil
}

// ViewsLastDays counts every project's views over the last days.
func ViewsLastDays(days int) int {
	var rows []struct{ Views int }
	_ = facades.Orm().Query().Raw(`SELECT COALESCE(sum(views), 0)::int AS views FROM page_views WHERE day > CURRENT_DATE - ?::int`, days).Scan(&rows)
	if len(rows) == 0 {
		return 0
	}
	return rows[0].Views
}
