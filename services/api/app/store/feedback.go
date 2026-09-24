package store

import (
	"strings"
	"sync"
	"time"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// MaxFeedbackMessage is the longest message a reader may send.
const MaxFeedbackMessage = 2000

// RateLimiter allows at most Limit events per key within Window. In memory:
// good enough to blunt a single spammer on one API process.
type RateLimiter struct {
	Limit  int
	Window time.Duration
	mu     sync.Mutex
	hits   map[string][]time.Time
}

// Allow records an event for key and reports whether it is within limits.
func (r *RateLimiter) Allow(key string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hits == nil {
		r.hits = map[string][]time.Time{}
	}
	cutoff := now.Add(-r.Window)
	kept := r.hits[key][:0]
	for _, t := range r.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= r.Limit {
		r.hits[key] = kept
		return false
	}
	r.hits[key] = append(kept, now)
	if len(r.hits) > 10000 { // forget everyone rather than grow unbounded
		r.hits = map[string][]time.Time{key: r.hits[key]}
	}
	return true
}

var feedbackLimiter = &RateLimiter{Limit: 20, Window: time.Hour}

// SubmitFeedback stores a reader's answer for a published page. Over the
// rate limit it silently drops the answer (accepted reports false).
func SubmitFeedback(p models.Project, slug string, helpful bool, message, ip string) (accepted bool, err error) {
	message = strings.TrimSpace(message)
	if len([]rune(message)) > MaxFeedbackMessage {
		return false, ValidationError{"message is too long (max 2000 characters)"}
	}
	pg, err := FindPage(p, slug)
	if err != nil {
		return false, err
	}
	if !feedbackLimiter.Allow(ip, time.Now()) {
		return false, nil
	}
	f := models.Feedback{ProjectID: p.ID, Slug: pg.Slug, Helpful: helpful, Message: message}
	return true, facades.Orm().Query().Create(&f)
}

type FeedbackEntry struct {
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Helpful   bool   `json:"helpful"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

type FeedbackTotal struct {
	Slug       string `json:"slug"`
	Helpful    int    `json:"helpful"`
	NotHelpful int    `json:"not_helpful"`
}

// ProjectFeedback lists up to 500 recent answers and per-page totals.
func ProjectFeedback(p models.Project) ([]FeedbackEntry, []FeedbackTotal, error) {
	var rows []models.Feedback
	if err := facades.Orm().Query().Where("project_id", p.ID).Order("id desc").Limit(500).Get(&rows); err != nil {
		return nil, nil, err
	}
	entries := []FeedbackEntry{}
	for _, f := range rows {
		e := FeedbackEntry{ID: f.ID, Slug: f.Slug, Helpful: f.Helpful, Message: f.Message}
		if f.CreatedAt != nil {
			e.CreatedAt = f.CreatedAt.ToIso8601String()
		}
		entries = append(entries, e)
	}
	var totals []FeedbackTotal
	err := facades.Orm().Query().Raw(`SELECT slug,
		count(*) FILTER (WHERE helpful) AS helpful,
		count(*) FILTER (WHERE NOT helpful) AS not_helpful
		FROM feedback WHERE project_id = ? GROUP BY slug ORDER BY slug`, p.ID).Scan(&totals)
	if totals == nil {
		totals = []FeedbackTotal{}
	}
	return entries, totals, err
}
