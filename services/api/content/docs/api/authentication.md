---
title: Authentication
description: Session tokens, API tokens and how they are checked.
position: 1
---

Admin endpoints and MCP write tools take a bearer token:

```http
Authorization: Bearer jd_4fb29…
```

## Two kinds of token

| Kind | Created by | Used for |
| ---- | ---------- | -------- |
| `session` | `POST /api/auth/login` | The admin app; revoked by `POST /api/auth/logout` |
| `api` | `POST /api/admin/tokens` (the admin's **Tokens** page) | Scripts and the [MCP server](/docs/mcp) |

Both are random 32-byte values prefixed with `jd_`. The plain token is
returned once; the database only stores its SHA-256 hash, plus `last_used_at`.

## Logging in

<Tabs items="curl,TypeScript">
<Tab value="curl">

```sh
curl -X POST https://api.jevidocs.jevido.app/api/auth/login \
  -H 'content-type: application/json' \
  -d '{"email":"admin@example.com","password":"…"}'
```

</Tab>
<Tab value="TypeScript">

```ts
const res = await fetch('https://api.jevidocs.jevido.app/api/auth/login', {
  method: 'POST',
  headers: { 'content-type': 'application/json' },
  body: JSON.stringify({ email, password }),
})
const { token, user } = await res.json()
```

</Tab>
</Tabs>

```json title="Response"
{ "token": "jd_…", "user": { "id": 1, "name": "admin", "email": "admin@example.com" } }
```

A wrong email or password answers `401` with `invalid email or password`;
which of the two is not revealed.

## The first admin

There is no sign-up. On start, when the `users` table is empty and both
`ADMIN_EMAIL` and `ADMIN_PASSWORD` are set, the API creates that admin with a
bcrypt-hashed password.

<Callout type="warn">
Treat API tokens like passwords: anyone holding one can create, change and
delete pages in every project.
</Callout>
