---
title: Introduction
description: jevidocs is a documentation framework built on Go, Goravel, Svelte 5 and PostgreSQL, with an admin app, a REST API and an MCP server.
position: 0
section: Get started
---

jevidocs is a documentation framework in the spirit of
[fumadocs](https://fumadocs.dev), rebuilt on a different stack: a **Goravel**
API in Go stores your pages in **PostgreSQL**, renders their Markdown on the
server, and serves them to a **Svelte 5** docs reader. An admin app edits them,
and an **MCP server** lets AI agents read, search and write them.

You are reading jevidocs right now: these pages live in the repository as
Markdown, are synced into the `jevidocs` project when the API starts, and are
rendered by the same reader every other project uses.

<Cards>
<Card title="Quick start" href="/docs/quick-start" description="Run the whole stack locally in a few minutes." />
<Card title="What is jevidocs?" href="/docs/what-is-jevidocs" description="How it maps to fumadocs, and where it differs." />
<Card title="Writing" href="/docs/writing" description="Markdown, front matter and the page tree." />
<Card title="Components" href="/docs/components" description="Callouts, cards, tabs, steps, accordions, file trees and code blocks." />
<Card title="REST API" href="/docs/api" description="Every public and admin endpoint." />
<Card title="MCP server" href="/docs/mcp" description="Let Claude and other agents read and write your docs." />
</Cards>

## What you get

- **A docs reader** at [jevidocs.jevido.app](https://jevidocs.jevido.app/docs):
  sidebar tree, search (<kbd>⌘</kbd> <kbd>K</kbd>), table of contents with
  scrollspy, breadcrumbs, previous/next links, light and dark themes, and
  "Copy Markdown" / "Open in ChatGPT or Claude" actions on every page.
- **An admin app** at
  [admin.jevidocs.jevido.app](https://admin.jevidocs.jevido.app) for projects,
  pages (with a live preview) and API tokens.
- **A REST API** at `https://api.jevidocs.jevido.app` that returns rendered
  pages, trees, search results and `llms.txt`.
- **An MCP server** at `https://api.jevidocs.jevido.app/mcp` with tools to list
  projects, read and search pages, and (with a token) create, update and delete
  them.

<Callout type="info" title="Many projects, one reader">
jevidocs hosts any number of documentation projects. The `jevidocs` project is
served at `/docs`; every other public project lives at `/p/<project>`.
</Callout>

## The stack

| Part | Technology |
| ---- | ---------- |
| API, MCP | Go 1.27, [Goravel](https://www.goravel.dev), `modelcontextprotocol/go-sdk` |
| Markdown | goldmark (CommonMark + GFM), Chroma highlighting |
| Database | PostgreSQL 18, full-text search with `tsvector` |
| Reader, admin | Svelte 5 (runes), Vite, TypeScript — no SvelteKit |
| Hosting | Coolify, one Containerfile per deployed unit |
