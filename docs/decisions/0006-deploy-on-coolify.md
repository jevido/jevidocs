# 0006. Three Coolify applications and a Postgres resource

- Status: accepted
- Date: 2026-09-24

## Context

Deploy to the existing Coolify on coolify.jevido.app from `main`, with the
docs site on jevidocs.jevido.app. fumadocs runs its docs and product on one
Next.js deployment; our units are separate.

## Decision

- Coolify project **jevidocs**: `jevidocs-site` (jevidocs.jevido.app),
  `jevidocs-admin` (admin.jevidocs.jevido.app), `jevidocs-api`
  (api.jevidocs.jevido.app) as Dockerfile applications built from the repo
  root with `infra/deploy/<resource>/Containerfile`, and `jevidocs-db`
  (Postgres 18) on the internal network.
- Watch paths per application; the GitHub App deploys on push to `main`.
- Secrets (APP_KEY, DB and admin credentials) exist only as Coolify
  environment variables.
- The zero-downtime pieces from jevido/work (drain script, asset carry
  forward, no-cache HTML) are reused.

## Consequences

- Cross-origin calls from site and admin to the API (CORS open for reads and
  bearer-token writes; no cookies).
- Three hostnames under the existing wildcard DNS; no DNS changes needed.

## Alternatives considered

- **Path-based routing on one domain:** fewer hostnames, but Coolify/Traefik
  path prefixes with rewriting add fragile proxy config.
