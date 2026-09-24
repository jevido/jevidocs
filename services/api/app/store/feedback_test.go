package store

import (
	"testing"
	"time"

	"dev.jevido/jevidocs/services/api/app/models"
)

func TestRateLimiter(t *testing.T) {
	r := &RateLimiter{Limit: 2, Window: time.Hour}
	now := time.Now()
	if !r.Allow("a", now) || !r.Allow("a", now) {
		t.Fatal("first two should pass")
	}
	if r.Allow("a", now) {
		t.Fatal("third should be limited")
	}
	if !r.Allow("b", now) {
		t.Fatal("other key limited")
	}
	if !r.Allow("a", now.Add(2*time.Hour)) {
		t.Fatal("window did not expire")
	}
}

func TestRevisionWorthy(t *testing.T) {
	a := models.Page{Title: "A", Body: "x"}
	b := a
	b.Position = 3
	if RevisionWorthy(a, b) {
		t.Error("position-only change should not make a revision")
	}
	b.Body = "y"
	if !RevisionWorthy(a, b) {
		t.Error("body change should make a revision")
	}
}

func TestRateLimiterExceededDoesNotRecord(t *testing.T) {
	r := &RateLimiter{Limit: 2, Window: time.Minute}
	now := time.Now()
	if r.Exceeded("a", now) {
		t.Fatal("fresh key exceeded")
	}
	r.Allow("a", now)
	r.Allow("a", now)
	if !r.Exceeded("a", now) || r.Exceeded("b", now) {
		t.Fatal("wrong exceeded state")
	}
	if r.Exceeded("a", now.Add(2*time.Minute)) {
		t.Fatal("window did not expire")
	}
}
