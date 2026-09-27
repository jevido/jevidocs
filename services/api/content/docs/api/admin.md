---
title: Admin endpoints
description: Managing projects, pages and tokens over the API.
position: 2
---

Every admin endpoint needs `Authorization: Bearer <token>`; see
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
