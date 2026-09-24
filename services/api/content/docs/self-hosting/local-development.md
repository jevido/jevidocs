---
title: Local development
description: Run and work on every unit on your machine.
position: 2
---

Requirements: Go 1.27+, Bun, Podman and [Task](https://taskfile.dev).
JavaScript dependencies are managed with **Bun only**.

## Ports

Local ports stay clear of framework defaults so jevidocs runs next to other
projects. Servers use `strictPort`, so a clash fails loudly.

| Port | Unit |
| ---- | ---- |
| 4720 | `apps/site` dev server (4721 preview) |
| 4730 | `services/api` |
| 4732 | PostgreSQL 18 from `infra/dev/compose.yml` |
| 4740 | `apps/admin` dev server (4741 preview) |

## Running things

<Tabs items="task,by hand">
<Tab value="task">

```sh
bun install
task db:up        # Postgres on 127.0.0.1:4732
task api:dev      # API on :4730, creates .env and APP_KEY on first run
task site:dev     # reader on :4720
task admin:dev    # admin on :4740
task check        # gofmt, go vet, tests, svelte-check
task down         # stop everything
```

</Tab>
<Tab value="by hand">

```sh
podman compose -f infra/dev/compose.yml up -d --wait
cd services/api && cp .env.example .env && go run . artisan key:generate && go run .
cd apps/site && bun run dev
cd apps/admin && bun run dev
```

</Tab>
</Tabs>

Point the frontends at the local API with `VITE_API_URL=http://127.0.0.1:4730`
(an `.env` file in the app directory works too).

## Editing these docs

The pages you are reading are Markdown files in `services/api/content/docs`,
embedded into the API binary. Restart the API to sync changes into the
`jevidocs` project.

<Files>
<Folder name="services/api/content/docs" defaultOpen>
<File name="index.md" />
<File name="quick-start.md" />
<Folder name="components">
<File name="index.md" />
<File name="callout.md" />
</Folder>
</Folder>
</Files>

## Tests

```sh
cd services/api && go test ./...
```

The Markdown renderer and page tree in `app/docs` are covered by unit tests
that need no database.
