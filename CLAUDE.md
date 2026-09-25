# CLAUDE.md

Guidance for coding agents (and people) working in this repository.

## What this is

**jevidocs** is a documentation framework modelled on fumadocs, built on Go
(Goravel), Svelte 5 and PostgreSQL. `services/api` stores projects and pages,
renders Markdown with fumadocs-style components, serves search, `llms.txt`
and the hosted MCP server. `apps/site` is the landing page and the docs
reader, `apps/admin` the editor. Read `README.md` for the overview and
`docs/decisions/` for why things are the way they are.

## Where things go

| Directory   | Rule                                                                  |
| ----------- | --------------------------------------------------------------------- |
| `apps/`     | A person opens it and interacts with it (websites).                   |
| `services/` | Runs unattended; no person interacts with it directly (APIs, MCP).    |
| `packages/` | Code used by **two or more** apps/services. Not before.               |
| `infra/`    | Podman/compose, Containerfiles, deploy config. No application code.   |
| `docs/`     | Why a decision was made. Never how code works.                        |

- Each unit is one directory directly under its category, e.g.
  `services/api`, `apps/site`. Keep a short `README.md` in each unit saying
  what it is and how to run it.
- A new Go unit gets its own `go.mod` (module `dev.jevido/jevidocs/<category>/<name>`)
  and is registered with `go work use ./<category>/<name>`.
- A new JS unit gets its own `package.json` (name `@jevidocs/<name>`); root Bun
  workspaces pick up `apps/*` and `packages/*`.
- Start code inside the unit that needs it. Move it to `packages/` only when a
  second consumer actually appears.

## Stack rules

- **Svelte 5 + Vite + TypeScript. Never SvelteKit.** Use runes (`$state`,
  `$derived`, `$effect`, `$props`); no legacy `export let` or stores for new
  code. Load the `svelte-code-writer` / `svelte-core-bestpractices` skills
  before writing `.svelte` files.
- **Go 1.27.** `gofmt` and `go vet` clean.
- **Goravel** for HTTP services in `services/`. Load the `goravel-development`
  skill (`.claude/skills/`). Services are JSON APIs only: no server-rendered
  views; UIs live in `apps/`.
- **Rendering happens in the API.** Markdown, components, TOC and search
  sections are produced by `services/api/app/docs` when a page is saved; the
  site only styles the HTML contract in `services/api/README.md` and attaches
  behaviour (tabs, copy buttons). Change the contract in both places.
  OpenAPI pages are the one exception: the API derives a structured
  reference (`app/openapi`) and the site renders it (`src/docs/api`); see
  `docs/decisions/0011-native-openapi-reference.md`.
- **Hosted MCP** lives in `services/api` (`app/mcpserver`, route `/mcp`).
  Tools are thin wrappers over `app/store`, the same code the REST routes use.
- **Bun only** for JS. Do not use npm, pnpm or yarn, and do not commit their
  lockfiles.
- **Podman**, not Docker, in scripts and docs (`podman compose`,
  `Containerfile`).

## Commands

```sh
bun install          # JS deps for every workspace
task                 # list tasks
task dev             # Postgres + API + site + admin
task check           # gofmt, go vet, go test, svelte-check across units
task down            # stop every dev server this repo started
```

**Local ports** use the 47xx range, one decade per unit, with `strictPort`:

| Port | Unit                             |
| ---- | -------------------------------- |
| 4720 | `apps/site` dev (4721 preview)   |
| 4730 | `services/api` HTTP              |
| 4732 | Local Postgres 18 (`task db:up`) |
| 4740 | `apps/admin` dev (4741 preview)  |

New units take the next free decade; add them here and to `DEV_PORTS` in the
root `Taskfile.yml` so `task down` stops them.

## Conventions

- **Commits:** conventional commits scoped by unit, e.g.
  `feat(site): ...`, `fix(api): ...`, `chore(infra): ...`, `docs: ...`.
- **`.gitignore`:** anchor patterns (`/bin/`, not `bin`).
- **Secrets:** never in the repo. Production values are Coolify environment
  variables; commit a `.env.example` when a unit needs config.
- **Deploy:** Coolify (https://coolify.jevido.app/) builds everything itself
  from pushes to `main`; there is no GitHub Actions. Each deployed unit is a
  Coolify application with `infra/deploy/<resource>/Containerfile`
  (+ `Containerfile.dockerignore`), built from the repo root, rebuilt only
  when its watch paths change. See `infra/deploy/README.md`.
- **Database:** PostgreSQL 18, a Coolify database resource in production and
  `infra/dev/compose.yml` locally. Migrations run when the API starts; write
  them to be safe alongside the previous version.
- **Own docs:** jevidocs' documentation is `services/api/content/docs/*.md`,
  synced into the `jevidocs` project on every API start. Edit the files, not
  the pages in the admin (those edits are overwritten on the next start).
- **Comments** explain non-obvious *how* and *why here*. Longer reasoning goes
  in a decision record.

## Docs discipline

`docs/` records **why**, not how. When you make a choice that is costly to
reverse (a dependency, a protocol, a data model, a deployment shape), write a
decision record in `docs/decisions/` using `0000-template.md`, numbered next
in sequence.

## Working style for agents

- Do what the task asks, in the unit it concerns.
- Verify before claiming done: build, test, and run the thing where possible.
  Say plainly what you could not verify.
- Prefer existing libraries and patterns already in the repo over new ones.
