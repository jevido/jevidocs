# deploy

Production runs on Coolify (https://coolify.jevido.app), project **jevidocs**,
environment **production**, on the single VPS. Coolify builds every image
itself from this repository; there is no GitHub Actions pipeline.

| Resource       | Domain                            | Containerfile                      | Port | Health     |
| -------------- | --------------------------------- | ---------------------------------- | ---- | ---------- |
| jevidocs-site  | https://jevidocs.jevido.app       | `infra/deploy/site/Containerfile`  | 80   | `/healthz` |
| jevidocs-admin | https://admin.jevidocs.jevido.app | `infra/deploy/admin/Containerfile` | 80   | `/healthz` |
| jevidocs-api   | https://api.jevidocs.jevido.app   | `infra/deploy/api/Containerfile`   | 4730 | `/health`  |
| jevidocs-db    | internal only                     | Coolify PostgreSQL 18              | 5432 | built in   |

All three are Coolify **applications** (Dockerfile build pack), sourced from
`jevido/jevidocs` on `main` through Coolify's GitHub App, built with the
repository root as context.

## When things deploy

A push to `main` deploys each resource whose **watch paths** changed:

- jevidocs-site: `apps/site/**`, `apps/cli/**`, `packages/**`, `infra/deploy/site/**`,
  `infra/deploy/shared/**`, `package.json`, `bun.lock`, `bunfig.toml`
- jevidocs-admin: `apps/admin/**`, `packages/**`, `infra/deploy/admin/**`,
  `infra/deploy/shared/**`, `package.json`, `bun.lock`, `bunfig.toml`
- jevidocs-api: `services/api/**` (including the docs in
  `services/api/content`), `infra/deploy/api/**`, `infra/deploy/shared/**`

Redeploy by hand from the Coolify UI, or with the API:

```sh
curl -X POST -H "Authorization: Bearer $COOLIFY_API_KEY" \
  "https://coolify.jevido.app/api/v1/deploy?uuid=<app uuid>"
```

| Resource       | Coolify UUID               |
| -------------- | -------------------------- |
| jevidocs-site  | `i6ohiq2cuw8vstzh11bipvmk` |
| jevidocs-admin | `fsalnpsnpewfm1g7scmkk9of` |
| jevidocs-api   | `iocc1mz3xp1hrdmw6n9ev3ra` |
| jevidocs-db    | `iy2oluigcdfr2qlijtknqpn3` (database) |

## What each image does

- **site** type-checks and builds `apps/site` with `VITE_API_URL` pointing
  at the API, and serves it with nginx. `/docs/*` and `/p/*` fall back to
  their single-page entries; `/llms.txt`, `/llms-full.txt` and
  `/docs/<slug>.md` redirect to the API. It also cross-compiles `apps/cli`
  and serves the binaries and `checksums.txt` from `/downloads/`.
- **admin** does the same for `apps/admin` (hash router, no fallbacks).
- **api** runs the tests, builds `services/api` and runs it as a non-root
  user. On start it migrates, creates the first admin from `ADMIN_EMAIL` /
  `ADMIN_PASSWORD` if there are no users, and syncs the embedded docs.

## Configuration

Coolify environment variables on jevidocs-api (runtime only): `APP_KEY` (32
characters), `APP_NAME`, `APP_URL`, `SITE_URL`, `ADMIN_EMAIL`,
`ADMIN_PASSWORD`, and `DB_HOST` / `DB_PORT` / `DB_DATABASE` / `DB_USERNAME` /
`DB_PASSWORD` for jevidocs-db (host is the database's UUID on Coolify's
internal network). Optional: `GITHUB_TOKEN` (sync private repos),
`ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL` (turn on Ask AI). The site and admin images take `VITE_API_URL` as a build
argument, defaulting to the production API.

## Database

jevidocs-db is a Coolify-managed PostgreSQL 18 resource (`postgres:18-alpine`),
not exposed publicly. Its volume is mounted on `/var/lib/postgresql`, the
path Postgres 18 needs. Coolify backs up the `jevidocs` database daily at
03:00 to `/data/coolify/backups/` on the VPS and keeps 7. The API migrates on
start and refuses to start if it cannot.

## Zero-downtime deploys

Same approach as jevido/work (its decision 0009): every image runs its server
under [`shared/drain.sh`](shared/drain.sh), the site and admin carry the
previous release's hashed assets forward (`_assets.txt`) and fall back to the
public host for a missing asset, and HTML is served `no-cache`. The optional
Traefik servers transport is in
[`traefik/jevidocs.yaml`](traefik/jevidocs.yaml); copy it to
`/data/coolify/proxy/dynamic/` on the server if the retry labels are added in
Coolify.

## Reproduce a build locally

```sh
task deploy:build:site
task deploy:build:admin
task deploy:build:api
```
