package store

import (
	"fmt"
	"net/url"
	"strings"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// LLMsTxt is the llms.txt index of a project (https://llmstxt.org): a title,
// a summary, and a link per page in reading order.
func LLMsTxt(p models.Project) (string, error) {
	tree, err := Tree(p)
	if err != nil {
		return "", err
	}
	descs := map[string]string{}
	pages, err := pagesOf(p, true)
	if err != nil {
		return "", err
	}
	for _, pg := range pages {
		descs[pg.Slug] = pg.Description
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", p.Name)
	if p.Description != "" {
		fmt.Fprintf(&b, "> %s\n\n", p.Description)
	}
	b.WriteString("## Docs\n\n")
	for _, l := range tree.Flatten() {
		fmt.Fprintf(&b, "- [%s](%s)", l.Title, MarkdownURL(p, l.Slug))
		if d := descs[l.Slug]; d != "" {
			fmt.Fprintf(&b, ": %s", d)
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

// MarkdownURL is where one page's Markdown is served by the API.
func MarkdownURL(p models.Project, slug string) string {
	base := strings.TrimRight(facades.Config().GetString("http.url"), "/")
	return fmt.Sprintf("%s/api/projects/%s/page.md?slug=%s", base, p.Slug, url.QueryEscape(slug))
}

// PageMarkdown is one page as Markdown with its title on top.
func PageMarkdown(pg models.Page, url string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", pg.Title)
	if pg.Description != "" {
		fmt.Fprintf(&b, "> %s\n\n", pg.Description)
	}
	if url != "" {
		fmt.Fprintf(&b, "Source: %s\n\n", url)
	}
	b.WriteString(strings.TrimSpace(pg.Body))
	b.WriteString("\n")
	return b.String()
}

// LLMsFull is every published page's Markdown, in reading order.
func LLMsFull(p models.Project) (string, error) {
	tree, err := Tree(p)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, l := range tree.Flatten() {
		pg, err := FindPage(p, l.Slug)
		if err != nil {
			continue
		}
		b.WriteString(PageMarkdown(pg, PageURL(p, pg.Slug)))
		b.WriteString("\n---\n\n")
	}
	return b.String(), nil
}
