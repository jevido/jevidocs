package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// dialect summarises jevidocs' Markdown for models writing pages.
const dialect = `jevidocs pages are CommonMark + GFM Markdown with front matter:

---
title: Page title
description: One sentence shown under the title and in search.
position: 10        # sort order among siblings
section: Guides     # top-level pages only: sidebar separator label
---

The slug is the page path (guides/install); a folder's page is its index.
Components go on their own lines, with blank lines around Markdown inside:

<Callout type="info|warn|error|success|idea" title="Optional">Text</Callout>
<Cards>
<Card title="Install" href="/docs/install" description="Get going" />
</Cards>
<Tabs items="npm,bun">
<Tab value="npm">...</Tab>
<Tab value="bun">...</Tab>
</Tabs>
<Steps><Step>

### First step

</Step></Steps>
<Accordions><Accordion title="Question">Answer</Accordion></Accordions>
<Files><Folder name="app" defaultOpen><File name="main.go" /></Folder></Files>

Code fences take a language and title="file.go"; {1,3-5} highlights lines.
GitHub alerts (> [!NOTE], [!TIP], [!WARNING]) and mermaid fences also work.
Use ## and ### headings: they build the table of contents.`

func registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "summarize_page",
		Title:       "Summarize a page",
		Description: "Summarise one documentation page, with its key points and links.",
		Arguments: []*mcp.PromptArgument{
			{Name: "project", Description: "Project slug, e.g. jevidocs", Required: true},
			{Name: "slug", Description: "Page slug; empty or index for the home page", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		a := req.Params.Arguments
		slug := a["slug"]
		if slug == "" {
			slug = "index"
		}
		text := fmt.Sprintf("Read the resource jevidocs://%s/%s (or call read_page with project %q and slug %q). "+
			"Summarise it in at most five bullet points for someone deciding whether to read it, "+
			"then list the headings it covers. Quote only what the page says.", a["project"], slug, a["project"], a["slug"])
		return &mcp.GetPromptResult{
			Description: "Summarise " + a["project"] + "/" + a["slug"],
			Messages:    []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: text}}},
		}, nil
	})

	server.AddPrompt(&mcp.Prompt{
		Name:        "write_page",
		Title:       "Write a page",
		Description: "Draft a new documentation page in jevidocs' Markdown dialect, then create it.",
		Arguments: []*mcp.PromptArgument{
			{Name: "project", Description: "Project slug", Required: true},
			{Name: "topic", Description: "What the page should explain", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		a := req.Params.Arguments
		text := fmt.Sprintf("Write a documentation page about: %s\n\n"+
			"First call get_page_tree for project %q and search_docs for related pages, so the new page fits "+
			"the structure, picks a sensible slug and links to existing pages instead of repeating them. "+
			"Then write it and save it with create_page (it needs an API token).\n\n%s", a["topic"], a["project"], dialect)
		return &mcp.GetPromptResult{
			Description: "Write a page about " + a["topic"],
			Messages:    []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: text}}},
		}, nil
	})
}
