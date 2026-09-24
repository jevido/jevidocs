---
title: Quick start
description: Run the API, the docs reader and the admin app on your machine.
position: 1
section: Get started
---

You need **Go 1.27+**, **Bun**, **Podman** (or Docker) and optionally
[Task](https://taskfile.dev).

<Steps>
<Step>

### Clone and install

```sh title="terminal"
git clone https://github.com/jevido/jevidocs
cd jevidocs
bun install
```

</Step>
<Step>

### Start PostgreSQL

The development database runs from `infra/dev/compose.yml` on `127.0.0.1:4732`.

<Tabs items="task,podman">
<Tab value="task">

```sh
task db:up
```

</Tab>
<Tab value="podman">

```sh
podman compose -f infra/dev/compose.yml up -d --wait
```

</Tab>
</Tabs>

</Step>
<Step>

### Run the API

```sh title="services/api"
cp .env.example .env
go run . artisan key:generate
go run .
```

The API migrates the database, creates the first admin from `ADMIN_EMAIL` /
`ADMIN_PASSWORD` (`admin@example.com` / `admin` in `.env.example`), syncs these
docs into the `jevidocs` project and listens on `http://127.0.0.1:4730`.

</Step>
<Step>

### Run the reader and the admin app

```sh
cd apps/site && VITE_API_URL=http://127.0.0.1:4730 bun run dev   # :4720
cd apps/admin && VITE_API_URL=http://127.0.0.1:4730 bun run dev  # :4740
```

Open [http://127.0.0.1:4720/docs](http://127.0.0.1:4720/docs) for the docs and
[http://127.0.0.1:4740](http://127.0.0.1:4740) for the admin.

</Step>
</Steps>

<Callout type="warn" title="Change the admin password">
The first admin is only created when the `users` table is empty. Set real
values for `ADMIN_EMAIL` and `ADMIN_PASSWORD` before the first start in any
shared environment.
</Callout>

## Next steps

<Cards>
<Card title="Write your first page" href="/docs/writing" description="Markdown, front matter and how the sidebar is built." />
<Card title="Connect an agent" href="/docs/mcp" description="Use the MCP server from Claude Code." />
</Cards>
