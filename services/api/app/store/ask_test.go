package store

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAskPromptHoldsExcerptsThenQuestion(t *testing.T) {
	got := askPrompt("Demo", "How do I install?", []askSnippet{{Title: "Install", URL: "https://x/install", Text: "Run bun install."}})
	for _, want := range []string{`title="Install"`, `url="https://x/install"`, "Run bun install.", "Question: How do I install?"} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt missing %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "Run bun install.") > strings.Index(got, "Question:") {
		t.Error("question must come after the excerpts")
	}
}

func TestCallClaudeAgainstFakeMessagesAPI(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("X-Api-Key") != "test-key" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-5-5",
			"content":[{"type":"text","text":"Run **bun install**. [Install](https://x/install)"}],
			"stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5}}`)
	}))
	defer srv.Close()

	got, err := callClaude(context.Background(), "test-key", srv.URL, "claude-opus-5-5", "Question: install?")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Run **bun install**. [Install](https://x/install)" {
		t.Errorf("answer = %q", got)
	}
	if body["model"] != "claude-opus-5-5" || body["system"] == nil {
		t.Errorf("request body = %v", body)
	}
}
