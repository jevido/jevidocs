package store

import (
	"testing"
	"time"
)

func TestPreviewToken(t *testing.T) {
	key := []byte("k")
	now := time.Unix(1_700_000_000, 0)
	tok := previewToken(key, 42, now.Add(time.Hour))
	if id, ok := checkPreviewToken(key, tok, now); !ok || id != 42 {
		t.Fatalf("valid token rejected: %v %v", id, ok)
	}
	if _, ok := checkPreviewToken(key, tok, now.Add(2*time.Hour)); ok {
		t.Fatal("expired token accepted")
	}
	if _, ok := checkPreviewToken([]byte("other"), tok, now); ok {
		t.Fatal("token from another key accepted")
	}
	if _, ok := checkPreviewToken(key, "43"+tok[2:], now); ok {
		t.Fatal("tampered page id accepted")
	}
	if _, ok := checkPreviewToken(key, "garbage", now); ok {
		t.Fatal("garbage accepted")
	}
}

func TestAssetSig(t *testing.T) {
	a, b := assetSig([]byte("k"), 1), assetSig([]byte("k"), 2)
	if a == b || a != assetSig([]byte("k"), 1) || len(a) != 32 {
		t.Fatalf("sigs %q %q", a, b)
	}
}
