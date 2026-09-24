package store

import (
	"testing"
	"time"
)

func TestHandoffCode(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	now := time.Now()
	code, err := handoffCode(key, 7, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := checkHandoffCode(key, code, now); !ok || id != 7 {
		t.Fatalf("valid code rejected: %d %v", id, ok)
	}
	if _, ok := checkHandoffCode(key, code, now.Add(2*time.Minute)); ok {
		t.Error("expired code accepted")
	}
	if _, ok := checkHandoffCode([]byte("another key another key another k"), code, now); ok {
		t.Error("code accepted with another key")
	}
	if _, ok := checkHandoffCode(key, "8"+code[1:], now); ok {
		t.Error("tampered user id accepted")
	}
}
