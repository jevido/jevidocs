package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// ErrAskDisabled is returned when no ANTHROPIC_API_KEY is configured.
var ErrAskDisabled = errors.New("ask is not enabled")

// askLimiter caps questions per client address: each one costs a model call.
var askLimiter = &RateLimiter{Limit: 20, Window: time.Hour}

// AskEnabled reports whether Ask AI is configured.
func AskEnabled() bool { return facades.Config().GetString("app.anthropic_key") != "" }

type AskSource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type AskResult struct {
	Answer  string      `json:"answer"`
	Sources []AskSource `json:"sources"`
}

// askSnippet is one piece of documentation given to the model.
type askSnippet struct {
	Title string
	URL   string
	Text  string
}

const askSystem = `You answer questions about one documentation site, using only the documentation excerpts provided in the user message.
- If the excerpts do not contain the answer, say so plainly and suggest what to search for; do not use outside knowledge.
- Answer in short Markdown: a direct answer first, then details or a code example if the excerpts have one.
- Cite the excerpts you used as Markdown links with their URL, e.g. [Quick start](https://...).`

// askPrompt builds the user message: the numbered excerpts, then the question.
func askPrompt(project, question string, snippets []askSnippet) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Documentation excerpts from %q:\n\n", project)
	for i, s := range snippets {
		fmt.Fprintf(&b, "<excerpt index=\"%d\" title=%q url=%q>\n%s\n</excerpt>\n\n", i+1, s.Title, s.URL, s.Text)
	}
	fmt.Fprintf(&b, "Question: %s", question)
	return b.String()
}

// askSnippets turns search results into excerpts: a heading's section text,
// or the start of a page. At most 6, each capped at 1500 characters.
func askSnippets(p models.Project, results []SearchResult) []askSnippet {
	var out []askSnippet
	seen := map[string]bool{}
	pages := map[string]models.Page{}
	for _, r := range results {
		key := r.Slug + "#" + r.Hash
		if seen[key] || len(out) >= 6 {
			continue
		}
		seen[key] = true
		pg, ok := pages[r.Slug]
		if !ok {
			var err error
			if pg, err = FindPage(p, r.Slug); err != nil {
				continue
			}
			pages[r.Slug] = pg
		}
		text := pg.Plain
		title := pg.Title
		if r.Hash != "" {
			var secs []docs.Section
			_ = json.Unmarshal([]byte(pg.Sections), &secs)
			for _, s := range secs {
				if s.ID == r.Hash {
					text, title = s.Text, pg.Title+" › "+s.Title
				}
			}
		}
		if len(text) > 1500 {
			text = text[:1500] + "…"
		}
		out = append(out, askSnippet{Title: title, URL: r.URL, Text: text})
	}
	return out
}

// Ask answers question from p's documentation with Claude.
func Ask(ctx context.Context, p models.Project, question, ip string) (AskResult, error) {
	key := facades.Config().GetString("app.anthropic_key")
	if key == "" {
		return AskResult{}, ErrAskDisabled
	}
	question = strings.TrimSpace(question)
	if len(question) < 3 || len(question) > 500 {
		return AskResult{}, ValidationError{"question must be 3 to 500 characters"}
	}
	if !askLimiter.Allow(ip, time.Now()) {
		return AskResult{}, ErrTooManyAttempts
	}
	results, err := Search(p, question, 20)
	if err != nil {
		return AskResult{}, err
	}
	snippets := askSnippets(p, results)
	res := AskResult{Sources: []AskSource{}}
	for _, s := range snippets {
		res.Sources = append(res.Sources, AskSource{Title: s.Title, URL: s.URL})
	}
	if len(snippets) == 0 {
		res.Answer = "I couldn't find anything about that in these docs. Try different words in search."
		return res, nil
	}
	cfg := facades.Config()
	answer, err := callClaude(ctx, key, cfg.GetString("app.anthropic_base_url"),
		cfg.GetString("app.anthropic_model", "claude-opus-5-5"), askPrompt(p.Name, question, snippets))
	if err != nil {
		facades.Log().Errorf("ask: %v", err)
		return AskResult{}, errors.New("the AI service did not answer, try again later")
	}
	res.Answer = answer
	return res, nil
}

func callClaude(ctx context.Context, key, baseURL, model, prompt string) (string, error) {
	opts := []option.RequestOption{option.WithAPIKey(key), option.WithMaxRetries(1)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	client := anthropic.NewClient(opts...)
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 1500,
		// Low effort keeps an answer inside the request timeout.
		OutputConfig: anthropic.OutputConfigParam{Effort: anthropic.OutputConfigEffortLow},
		System:       []anthropic.TextBlockParam{{Text: askSystem}},
		Messages:     []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	})
	if err != nil {
		return "", err
	}
	if resp.StopReason == "refusal" {
		return "I can't help with that question.", nil
	}
	var b strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(t.Text)
		}
	}
	return strings.TrimSpace(b.String()), nil
}
