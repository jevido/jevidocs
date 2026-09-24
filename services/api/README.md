# api

The jevidocs API: a Goravel JSON API over PostgreSQL that stores documentation
projects and their pages, renders Markdown (with fumadocs-style components) to
HTML, builds page trees and tables of contents, serves search, `llms.txt`, and
hosts the MCP server at `/mcp`.

Scaffolded from the Goravel scaffold (framework v1.18.0), trimmed to a JSON
API. No views: every UI lives in `apps/`.

```sh
task api:dev     # from the repo root; creates .env and APP_KEY on first run
task api:test
go run . artisan list
```

Module: `dev.jevido/jevidocs/services/api`. Local port 4730.

Live at https://api.jevidocs.jevido.app, built and deployed by Coolify from
`infra/deploy/api/`.

## Content

On start the API syncs the Markdown under `content/` (embedded in the binary)
into the `jevidocs` project, so the framework's own documentation is always
the one in the repository. Other projects are created and edited in the admin
app or over MCP.

Markdown is CommonMark + GFM, with YAML-ish front matter (`title`,
`description`, `icon`, `position`, `section`) and JSX-style components on
their own lines:

```md
<Callout type="warn" title="Heads up">
Body with **markdown**.
</Callout>

<Cards>
<Card title="Install" href="/docs/install" description="Get going" />
</Cards>

<Tabs items="npm,bun">
<Tab value="npm">...</Tab>
<Tab value="bun">...</Tab>
</Tabs>

<Steps>
<Step>

### First

</Step>
</Steps>

<Accordions><Accordion title="Why?">...</Accordion></Accordions>

<Files><Folder name="app" defaultOpen><File name="main.go" /></Folder></Files>
```

## HTTP API

All responses are JSON unless noted. Errors: `{"error": "message"}` with a
4xx/5xx status.

### Public

| Method | Path | Returns |
| ------ | ---- | ------- |
| GET | `/health` | `{status, database, framework}` |
| GET | `/api/projects` | `Project[]` (public projects) |
| GET | `/api/projects/{project}` | `Project & {tree: PageTree, versions?: VersionLink[]}` |
| GET | `/api/projects/{project}/page?slug=a/b` | `Page` (slug `""` = index) |
| GET | `/api/projects/{project}/search?q=term` | `SearchResult[]` (max 20) |
| GET | `/api/projects/{project}/llms.txt` | text/plain index of pages |
| GET | `/api/projects/{project}/llms-full.txt` | text/plain, every page's Markdown |
| GET | `/api/projects/{project}/page.md?slug=a/b` | text/markdown of one page |

Public reads (`/api/projects/{project}`, `page`, `page.md`, `search`, `llms.txt`,
`llms-full.txt`, `sitemap.xml`) take `?locale=nl` for projects with several
languages; a missing translation falls back to the default-locale page and
`Page.fallback` is `true`. `Project` includes `locales` (default first),
`default_locale` and the request's `locale`. Admin page inputs take `locale`
(create only; `''` = default) and projects take `locales` (comma list) and
`default_locale`.
| POST | `/api/projects/{project}/ask` | `{question}` → `{answer, sources: {title, url}[]}`; 404 when Ask AI is off (no `ANTHROPIC_API_KEY`) |
| POST | `/api/hooks/github/{project}` | GitHub push webhook (HMAC `X-Hub-Signature-256` with the project's source secret) → 202 |
| GET | `/api/assets/{id}/{name}` | the asset's bytes (immutable cache, `nosniff`, CSP on SVG) |
| POST | `/api/projects/{project}/views` | `{slug}` (any content type, parsed as JSON) → `{ok: true}`; bots, unknown pages and >300/h per IP are ignored |
| POST | `/api/projects/{project}/feedback` | `{slug, helpful, message?}` → `{ok: true}` (message ≤ 2000 chars; 20 per IP per hour) |

```ts
type Project = {
  slug: string; name: string; description: string;
  github_url: string; // may be ""
  links: { text: string; url: string }[]; // navbar links
  public: boolean; updated_at: string;
  version_group: string; version_label: string; // "" when unversioned
}

type VersionLink = { slug: string; name: string; label: string; url: string } // same version_group, newest label first

type PageTree = { name: string; children: TreeNode[] }
type TreeNode =
  | { type: 'page'; name: string; slug: string; icon?: string }
  | { type: 'separator'; name: string }
  | { type: 'folder'; name: string; icon?: string; index?: { type: 'page'; name: string; slug: string };
      children: TreeNode[]; defaultOpen: boolean }

type Page = {
  slug: string; title: string; description: string; icon: string;
  html: string;                      // rendered body, see "Rendered HTML"
  toc: { depth: number; title: string; url: string }[]; // url = "#heading-id", depth 2..4
  breadcrumbs: { name: string; slug?: string }[];        // folders + page
  previous: { title: string; slug: string } | null;
  next: { title: string; slug: string } | null;
  markdown: string;                  // source without front matter
  updated_at: string;
}

type SearchResult = {
  type: 'page' | 'heading';
  slug: string;          // page slug
  hash: string;          // "" or "heading-id"
  title: string;         // page or heading title
  page_title: string;
  snippet: string;       // HTML, matches wrapped in <mark>
  fuzzy?: true;          // trigram "similar result" when nothing matched exactly
}
```

### Admin (header `Authorization: Bearer <token>`)

| Method | Path | Body → Returns |
| ------ | ---- | -------------- |
| POST | `/api/auth/login` | `{email, password}` → `{token, user}` |
| GET | `/api/auth/me` | → `{user}` |
| POST | `/api/auth/logout` | → `{ok: true}` (revokes the token) |
| PUT | `/api/auth/password` | `{current, new}` → `{ok: true}` (new: 8–72 chars) |
| GET | `/api/admin/projects` | → `Project[]` (all, incl. private) |
| POST | `/api/admin/projects` | `{slug, name, description, github_url, links, public}` → `Project` |
| PUT | `/api/admin/projects/{project}` | same fields → `Project` |
| DELETE | `/api/admin/projects/{project}` | → `{ok: true}` |
| PUT | `/api/admin/projects/{project}/order` | `{items: [{id, position}]}` → `{ok: true}` (one transaction; only that project's pages) |
| PUT | `/api/admin/projects/{project}/sync` | `{files: {"guides/index.md": "..."}, prune}` → `{created, updated, unchanged, deleted}` |
| GET | `/api/admin/projects/{project}/source` | → `{repo, ref, path, secret, synced_at, status, webhook_url}` (GitHub source) |
| PUT | `/api/admin/projects/{project}/source` | `{repo, ref, path}` → same; empty `repo` disconnects |
| POST | `/api/admin/projects/{project}/source/sync` | → `202 {started}`; poll `GET .../source` for `status` / `synced_at` |
| GET | `/api/admin/projects/{project}/pages` | → `AdminPage[]` without `body` |
| POST | `/api/admin/projects/{project}/pages` | `AdminPageInput` → `AdminPage` |
| GET | `/api/admin/projects/{project}/pages/{id}` | → `AdminPage` |
| PUT | `/api/admin/projects/{project}/pages/{id}` | `AdminPageInput` → `AdminPage` |
| DELETE | `/api/admin/projects/{project}/pages/{id}` | → `{ok: true}` |
| POST | `/api/admin/preview` | `{body}` → `{html, toc}` |
| GET | `/api/admin/tokens` | → `{id, name, last_used_at, created_at}[]` |
| POST | `/api/admin/tokens` | `{name}` → `{id, name, token}` (plain token shown once) |
| DELETE | `/api/admin/tokens/{id}` | → `{ok: true}` |
| GET | `/api/admin/stats` | → `{projects, pages, tokens, views_30d}` |
| GET | `/api/admin/projects/{project}/assets` | → `Asset[]` newest first |
| POST | `/api/admin/projects/{project}/assets` | multipart `file` (≤ 8 MB; png, jpeg, gif, webp, svg, avif, pdf, txt) → `Asset` (same bytes → existing asset) |
| DELETE | `/api/admin/projects/{project}/assets/{id}` | → `{ok: true}` |
| GET | `/api/admin/users` | → `{id, name, email, created_at}[]` |
| POST | `/api/admin/users` | `{name, email, password}` → user |
| DELETE | `/api/admin/users/{id}` | → `{ok: true}` (not yourself; revokes their tokens) |

`Asset = {id, name, url, content_type, size, created_at}`; embed it with
`![name](url)`.
| GET | `/api/admin/projects/{project}/insights?days=30` | → `{days, views_total, views_by_day: {day, views}[], top_pages: {slug, title, views}[], top_searches: {query, count, results}[], zero_result_searches: {query, count, results}[], feedback: {helpful, not_helpful}}` |
| GET | `/api/admin/projects/{project}/feedback` | → `{entries: {id, slug, helpful, message, created_at}[], totals: {slug, helpful, not_helpful}[]}` (newest first) |
| GET | `/api/admin/projects/{project}/pages/{id}/revisions` | → `{id, title, created_at, size}[]` (newest first, last 50 kept) |
| GET | `/api/admin/projects/{project}/pages/{id}/revisions/{rid}` | → `{id, title, description, body, created_at, size}` |
| POST | `/api/admin/projects/{project}/pages/{id}/revisions/{rid}/restore` | → `AdminPage` (the replaced state becomes a revision) |

```ts
type User = { id: number; name: string; email: string }
type AdminPageInput = {
  slug: string; title: string; description: string; icon: string;
  position: number; section: string; published: boolean; body: string; // Markdown
}
type AdminPage = AdminPageInput & { id: number; project: string; updated_at: string; created_at: string }
```

Roles: `viewer` (GET only, plus preview and own tokens/password), `editor`
(also pages, assets, imports, syncs, order; MCP write tools), `admin` (also
project create/settings/delete, GitHub source settings, users). Enforced by
`middleware.Role()` via `store.RequiredRole(method, path)`; `PUT
/api/admin/users/{id}` `{role}` changes a role, and the last admin cannot be
demoted or deleted. Admin page views include `updated_by` (editor's name).

The first admin comes from `ADMIN_EMAIL` / `ADMIN_PASSWORD` on start when no
user exists.

### Page tree rules

- A page's slug is its path (`guides/install`); `""` is the project index.
- A folder is any slug prefix with children. A page whose slug equals the
  folder path (`guides`) is the folder's index and gives it its name and icon.
- Siblings sort by `position`, then title. A folder sorts by its index page's
  position (or its lowest child's).
- A top-level page with a non-empty `section` different from the previous
  sibling's starts a separator with that name.

### Rendered HTML

The server emits final HTML; clients only style it and attach behaviour.

- Headings: `<h2 id="x"><a class="fd-anchor" href="#x">Title</a></h2>`
- Code:
  `<figure class="fd-codeblock" data-lang="go"><figcaption class="fd-codeblock-title">main.go</figcaption><pre class="chroma"><code>…</code></pre></figure>`
  (Chroma CSS classes; figcaption only with `title="..."` in the fence info.)
- Callout:
  `<div class="fd-callout" data-type="info|warn|error|success|idea"><div class="fd-callout-title">…</div><div class="fd-callout-body">…</div></div>`
- Cards:
  `<div class="fd-cards"><div class="fd-card-slot"><a class="fd-card" href="…"><div class="fd-card-title">…</div><div class="fd-card-desc">…</div></a></div></div>` (the slot keeps the line a block-level HTML block; style it `display: contents`)
- Tabs:
  `<div class="fd-tabs"><div class="fd-tabs-list" role="tablist"><button class="fd-tab-trigger" data-tab="npm" data-active>npm</button>…</div><div class="fd-tab" data-value="npm" data-active>…</div>…</div>`
- Steps: `<div class="fd-steps"><div class="fd-step">…</div></div>`
- Accordions:
  `<div class="fd-accordions"><details class="fd-accordion"><summary>Title</summary><div class="fd-accordion-body">…</div></details></div>`
- Files:
  `<div class="fd-files"><details class="fd-folder" open><summary>app</summary><div class="fd-folder-body"><div class="fd-file">main.go</div></div></details></div>`
- Tables: plain `<table>`; task lists, strikethrough, autolinks per GFM.

## MCP

Streamable HTTP at `/mcp`, stateless, JSON responses. Tools:

- `list_projects`, `get_page_tree`, `read_page`, `search_docs`, `get_llms_txt`, `list_versions`: public.
- Resource templates `jevidocs://{project}/{slug}` (page Markdown) and `jevidocs://{project}/llms.txt`; prompts `summarize_page`, `write_page`.
- `create_page`, `update_page`, `delete_page`: need `Authorization: Bearer
  <token>` from an admin API token.

MCP tool calls must finish within Goravel's `http.request_timeout`: its
global timeout middleware buffers responses, which is also why the transport
answers with JSON rather than an event stream.
