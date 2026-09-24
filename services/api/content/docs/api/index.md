---
title: REST API
description: Every endpoint of the jevidocs API.
position: 30
section: Reference
root: true
icon: code
---

Base URL: `https://api.jevidocs.jevido.app` (locally `http://127.0.0.1:4730`).
Responses are JSON unless noted. Errors look like `{"error": "message"}` with
a `4xx`/`5xx` status: `404` for unknown or private projects and pages, `422`
for validation problems, `401` without a valid token. CORS allows every origin.

## Public

| Method | Path | Returns |
| ------ | ---- | ------- |
| GET | `/health` | `{status, database, framework}`; `503` when Postgres is unreachable |
| GET | `/api/projects` | Public projects |
| GET | `/api/projects/{project}` | Project with its page `tree` |
| GET | `/api/projects/{project}/page?slug=a/b` | A rendered page (`slug=` for the index) |
| GET | `/api/projects/{project}/search?q=term` | Up to 20 search results |
| GET | `/api/projects/{project}/llms.txt` | `text/plain` index of pages |
| GET | `/api/projects/{project}/llms-full.txt` | `text/plain`, every page's Markdown |
| GET | `/api/projects/{project}/page.md?slug=a/b` | `text/markdown` of one page |

<Tabs items="curl,TypeScript">
<Tab value="curl">

```sh
curl 'https://api.jevidocs.jevido.app/api/projects/jevidocs/page?slug=quick-start'
```

</Tab>
<Tab value="TypeScript">

```ts
const res = await fetch(
  'https://api.jevidocs.jevido.app/api/projects/jevidocs/page?slug=quick-start',
)
const page: Page = await res.json()
```

</Tab>
</Tabs>

```json title="Page"
{
  "slug": "quick-start",
  "title": "Quick start",
  "description": "Run the API, the docs reader and the admin app on your machine.",
  "icon": "",
  "html": "<figure class=\"fd-codeblock\" …",
  "toc": [{ "depth": 2, "title": "Next steps", "url": "#next-steps" }],
  "breadcrumbs": [{ "name": "Quick start", "slug": "quick-start" }],
  "previous": { "title": "Introduction", "slug": "" },
  "next": { "title": "What is jevidocs?", "slug": "what-is-jevidocs" },
  "markdown": "You need **Go 1.27+** …",
  "url": "https://jevidocs.jevido.app/docs/quick-start",
  "updated_at": "2026-09-24T13:12:28+00:00"
}
```

```json title="Project"
{
  "slug": "jevidocs",
  "name": "jevidocs",
  "description": "Documentation framework on Go, Goravel, Svelte 5 and PostgreSQL …",
  "github_url": "https://github.com/jevido/jevidocs",
  "links": [{ "text": "Docs", "url": "/docs" }],
  "public": true,
  "managed": true,
  "updated_at": "2026-09-24T13:12:28+00:00"
}
```

## Admin

All admin endpoints need `Authorization: Bearer <token>`; see
[Authentication](/docs/api/authentication).

| Method | Path | Body → Returns |
| ------ | ---- | -------------- |
| POST | `/api/auth/login` | `{email, password}` → `{token, user}` |
| GET | `/api/auth/me` | → `{user}` |
| POST | `/api/auth/logout` | → `{ok: true}`, revokes the token |
| GET | `/api/admin/stats` | → `{projects, pages, tokens}` |
| GET | `/api/admin/projects` | → every project, including private ones |
| POST | `/api/admin/projects` | `{slug, name, description, github_url, links, public}` → project |
| GET | `/api/admin/projects/{project}` | → project |
| PUT | `/api/admin/projects/{project}` | same fields → project |
| DELETE | `/api/admin/projects/{project}` | → `{ok: true}`, deletes its pages too |
| GET | `/api/admin/projects/{project}/pages` | → pages without `body` |
| POST | `/api/admin/projects/{project}/pages` | page input → page |
| GET | `/api/admin/projects/{project}/pages/{id}` | → page with `body` |
| PUT | `/api/admin/projects/{project}/pages/{id}` | page input → page |
| DELETE | `/api/admin/projects/{project}/pages/{id}` | → `{ok: true}` |
| POST | `/api/admin/preview` | `{body}` → `{html, toc}` |
| GET | `/api/admin/tokens` | → `{id, name, last_used_at, created_at}[]` |
| POST | `/api/admin/tokens` | `{name}` → `{id, name, token}` |
| DELETE | `/api/admin/tokens/{id}` | → `{ok: true}` |

```json title="Page input"
{
  "slug": "guides/install",
  "title": "Install",
  "description": "Get it running.",
  "icon": "",
  "position": 1,
  "section": "",
  "published": true,
  "body": "## Requirements\n\nGo 1.27 …"
}
```

<Callout type="info" title="Validation">
Page slugs are lowercase path segments (`guides/install`) or empty for the
index, and unique per project. `title` is required, unless the body's front
matter has one. Project slugs are lowercase letters, digits and dashes.
</Callout>
