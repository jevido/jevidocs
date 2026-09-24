# 0001. Monorepo layout from jevido/work

- Status: accepted
- Date: 2026-09-24

## Context

jevidocs needs a docs site, an admin, an API with MCP and a CLI. They share
one data model and one deployment target. jevido/work already settled a
monorepo shape (apps/, services/, packages/, infra/, docs/) that works with
Coolify's per-application watch paths.

## Decision

Use the same layout: `apps/site`, `apps/admin`, `apps/cli`, `services/api`,
`infra/dev`, `infra/deploy/<resource>`, `docs/decisions`. Bun workspaces for
JS, `go.work` for Go, one module per unit, Task as the runner.

## Consequences

- Each unit deploys on its own when its paths change.
- Conventions (ports, commits, Containerfiles, drain script) carry over, so
  the two repos read the same.
- No `packages/` yet: site and admin each keep their own copy of the rendered
  HTML styles until a shared package earns its place.

## Alternatives considered

- **fumadocs' own layout** (pnpm + turbo, packages published to npm): built
  for a library ecosystem; jevidocs is a hosted product first.
- **One app serving everything:** simpler deploy, but couples the reader's
  caching and release cadence to the admin's.
