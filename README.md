# jevidocs

A documentation framework in the spirit of [fumadocs](https://fumadocs.dev),
rebuilt on **Go (Goravel)**, **Svelte 5** and **PostgreSQL**. Pages live in a
database instead of files, are edited in an admin app or by agents over MCP,
and are read on a fumadocs-style docs site.

- **Docs site:** sidebar page tree with root-folder tabs, search (⌘K) with
  heading results, table of contents with scrollspy, breadcrumbs,
  previous/next (`[` / `]`), light/dark theme and accent presets, copy
  Markdown, "open in ChatGPT/Claude", edit-this-page links, page feedback,
  mobile TOC, print styles, a project directory at `/p`.
- **Markdown:** CommonMark + GFM, front matter, the fumadocs components
  (Callout, Cards, Tabs, Steps, Accordion, Files), GitHub alerts, code
  blocks with titles, highlighted lines, line numbers, diff/focus
  notations, Mermaid diagrams and zoomable images. Rendered in Go.
- **Admin:** projects (theme, banner, navbar links), pages with a live
  preview, component snippets and paste/drop image upload, revision
  history with diff and restore, assets, feedback, insights (views, top and
  zero-result searches), OpenAPI import, users and API tokens.
- **API:** public read endpoints (projects, trees, pages, search,
  `llms.txt`, `llms-full.txt`, per-page Markdown, sitemap) and authenticated
  admin endpoints, including bulk sync from files.
- **MCP:** hosted at `/mcp`; agents list, read and search docs, and with a
  token create, update and delete pages.
- **OpenAPI:** generate an API reference (a page per operation) from an
  OpenAPI 3 spec.
- **CLI:** `jevidocs init | push | pull | search | openapi` for docs as code;
  binaries at https://jevidocs.jevido.app/downloads/.

jevidocs documents itself: the Markdown under
[`services/api/content`](services/api/content) is synced into the `jevidocs`
project on every API start.

## Live

| What | Where |
| ---- | ----- |
| Landing page + docs | https://jevidocs.jevido.app (docs at [/docs](https://jevidocs.jevido.app/docs)) |
| Admin | https://admin.jevidocs.jevido.app |
| API + MCP | https://api.jevidocs.jevido.app (`/health`, `/api/...`, `/mcp`) |

Coolify builds and deploys every unit from pushes to `main`; see
[`infra/deploy`](infra/deploy).

## Layout

Follows the [jevido/work](https://github.com/jevido/work) monorepo layout.

| Directory   | What goes there                                                        |
| ----------- | ---------------------------------------------------------------------- |
| `apps/`     | Things a person opens: `site` (landing + docs reader), `admin`, `cli`.  |
| `services/` | Things that run unattended: `api` (Goravel REST API + hosted MCP).      |
| `packages/` | Code shared by two or more units (none yet).                            |
| `infra/`    | Podman compose for development, Containerfiles and deploy config.       |
| `docs/`     | Why decisions were made. Not how the code works.                        |

### How it maps to fumadocs

| fumadocs | jevidocs |
| -------- | -------- |
| `fumadocs-core` (source loader, page tree, TOC) | `services/api/app/docs` (Go) |
| `fumadocs-mdx` (content collections) | Markdown pages in Postgres, `content/` sync |
| `fumadocs-ui` (DocsLayout, components) | `apps/site` (Svelte 5) |
| Orama search | Postgres full-text search |
| `llms.txt` / `.mdx` routes | `/api/projects/{p}/llms.txt`, `/docs/<slug>.md` |
| MCP / AI integrations | Hosted MCP server at `/mcp` |
| — | `apps/admin` for editing content |

## Stack

- **Frontend:** Svelte 5 with Vite and TypeScript. No SvelteKit.
- **Backend:** Go 1.27 with [Goravel](https://www.goravel.dev), PostgreSQL 18.
- **Markdown:** goldmark (GFM) + Chroma highlighting, components expanded
  server-side.
- **Tooling:** Bun workspaces, a `go.work` workspace,
  [Task](https://taskfile.dev).
- **Containers:** Podman locally; Coolify on one VPS in production.

## Getting started

Requirements: Go 1.27+, Bun, Task, Podman.

```sh
bun install
task dev        # Postgres + API (4730) + site (4720) + admin (4740)
```

Sign in to the admin at http://127.0.0.1:4740 with `admin@example.com` /
`admin` (from `services/api/.env.example`).

Conventions for people and coding agents live in [`CLAUDE.md`](CLAUDE.md).
