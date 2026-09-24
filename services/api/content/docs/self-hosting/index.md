---
title: Self-hosting
description: What a jevidocs deployment consists of.
position: 40
section: Operate
---

A deployment is three container images and one PostgreSQL database. Each
image is built from the repository root with its own Containerfile in
`infra/deploy/<unit>/`.

| Unit | Image | Serves | Port | Health |
| ---- | ----- | ------ | ---- | ------ |
| API | `infra/deploy/api/Containerfile` | REST API and `/mcp` | 4730 | `/health` |
| Site | `infra/deploy/site/Containerfile` | Landing page and docs reader (nginx) | 80 | `/healthz` |
| Admin | `infra/deploy/admin/Containerfile` | Admin app (nginx) | 80 | `/healthz` |
| Database | PostgreSQL 18 | — | 5432 | built in |

<Cards>
<Card title="Deploy on Coolify" href="/docs/self-hosting/coolify" description="How jevidocs.jevido.app is deployed." />
<Card title="Local development" href="/docs/self-hosting/local-development" description="Task, Podman and the 47xx ports." />
</Cards>

## API configuration

The API is configured with environment variables; nothing secret is baked into
the image.

| Variable | Purpose |
| -------- | ------- |
| `APP_KEY` | 32-character application key |
| `APP_URL` | Public URL of the API, used in `llms.txt` links |
| `SITE_URL` | Public URL of the reader, used for page links |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | First admin, created when no user exists |
| `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD` | PostgreSQL connection |

The image already sets `APP_ENV=production`, `APP_HOST=0.0.0.0`,
`APP_PORT=4730`, `DB_CONNECTION=postgres` and `DB_MIGRATE_ON_START=true`.

## On start

<Steps>
<Step>

### Migrate

Pending migrations run before the server listens. If they fail, the process
exits, so a broken deploy never goes live.

</Step>
<Step>

### Prepare

The first admin is created if needed, and `services/api/content/docs` is
synced into the `jevidocs` project.

</Step>
<Step>

### Serve

`/health` pings the database and answers `503` when it cannot reach it.

</Step>
</Steps>

## Reader and admin configuration

Both frontends read the API address at build time from the `VITE_API_URL`
build argument (default `https://api.jevidocs.jevido.app`).
