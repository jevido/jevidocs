---
title: Architecture
description: The monorepo layout and how the units talk to each other.
position: 3
section: Get started
---

jevidocs is a monorepo with the same layout as
[jevido/work](https://github.com/jevido/work).

| Directory | What goes there |
| --------- | --------------- |
| `apps/` | Things a person opens: the docs reader (`site`) and the `admin` app. |
| `services/` | Things that run unattended: the Goravel `api`, including the MCP server. |
| `packages/` | Code shared by two or more units (none yet). |
| `infra/` | Compose files for development, Containerfiles and deploy config. |
| `docs/` | Decision records: why things are the way they are. |

<Files>
<Folder name="apps" defaultOpen>
<Folder name="site">
<File name="index.html" />
<File name="docs/index.html" />
<File name="src/docs/DocsApp.svelte" />
</Folder>
<Folder name="admin">
<File name="src/pages/PageEditor.svelte" />
</Folder>
</Folder>
<Folder name="services" defaultOpen>
<Folder name="api" defaultOpen>
<File name="app/docs/ (Markdown, tree)" />
<File name="app/store/ (database layer)" />
<File name="app/mcpserver/ (MCP tools)" />
<File name="app/http/controllers/" />
<File name="content/docs/ (these pages)" />
<File name="routes/web.go" />
</Folder>
</Folder>
<Folder name="infra">
<File name="dev/compose.yml" />
<File name="deploy/api/Containerfile" />
<File name="deploy/site/Containerfile" />
<File name="deploy/admin/Containerfile" />
</Folder>
</Files>

## Request flow

1. The reader loads `/api/projects/{project}` once for the navbar and the tree.
2. Each navigation fetches `/api/projects/{project}/page?slug=...`, which
   returns ready HTML, the TOC, breadcrumbs and previous/next links.
3. The reader styles the HTML and attaches behaviour: tab switching, copy
   buttons, anchor links and client-side navigation.

## One layer for REST and MCP

Controllers and MCP tools both call `app/store`, which validates input and
calls the pure `app/docs` package for rendering and trees. A tool and its REST
route cannot disagree about validation.

<Callout type="idea">
`app/docs` has no database or framework dependency, which keeps the Markdown
renderer and the tree builder easy to unit test.
</Callout>
