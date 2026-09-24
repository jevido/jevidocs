---
title: Coolify
description: How jevidocs is built and deployed on Coolify.
position: 1
---

jevidocs runs on [Coolify](https://coolify.io) on a single VPS. Coolify builds
every image itself from pushes to `main`; there is no CI pipeline.

| Resource | Domain | Containerfile |
| -------- | ------ | ------------- |
| Site | https://jevidocs.jevido.app | `infra/deploy/site/Containerfile` |
| Admin | https://admin.jevidocs.jevido.app | `infra/deploy/admin/Containerfile` |
| API | https://api.jevidocs.jevido.app | `infra/deploy/api/Containerfile` |
| Database | internal only | Coolify PostgreSQL 18 |

## Setting it up

<Steps>
<Step>

### Create the database

Add a PostgreSQL 18 database resource. Keep it private; the API reaches it on
Coolify's internal network, with the resource's UUID as hostname.

</Step>
<Step>

### Create three applications

For each unit, add an application from the GitHub repository on branch `main`
with the **Dockerfile** build pack, base directory `/` and the unit's
Containerfile path, for example `/infra/deploy/api/Containerfile`. Set the
port (4730 for the API, 80 for site and admin) and the domain.

</Step>
<Step>

### Configure the API

Set `APP_KEY`, `APP_URL`, `SITE_URL`, `ADMIN_EMAIL`, `ADMIN_PASSWORD` and the
`DB_*` variables as runtime environment variables.

</Step>
<Step>

### Deploy

Deploy each application. The API migrates, creates the admin and syncs the
docs; the site and admin are static nginx images.

</Step>
</Steps>

## Watch paths

A push to `main` triggers a deploy of each application whose watch paths
changed:

- Site: `apps/site/**`, `infra/deploy/site/**`, `package.json`, `bun.lock`
- Admin: `apps/admin/**`, `infra/deploy/admin/**`, `package.json`, `bun.lock`
- API: `services/api/**`, `infra/deploy/api/**`

## Zero-downtime deploys

<Accordions>
<Accordion title="Drain before stopping">
Every image runs its server under `infra/deploy/shared/drain.sh`: after a stop
request the container keeps serving for 5 seconds before its graceful
shutdown, so the proxy has switched to the new container first.
</Accordion>
<Accordion title="Assets from both releases">
The site and admin images carry the previous release's hashed assets forward
(`_assets.txt`), and nginx falls back to the public site for an asset it does
not have. HTML is served `no-cache`.
</Accordion>
<Accordion title="Compatible migrations">
For a few seconds old and new API containers serve side by side, so
migrations must work with the previous release: add first, remove later.
</Accordion>
</Accordions>

## Build locally

```sh
podman build -f infra/deploy/api/Containerfile -t jevidocs-api:local .
podman build -f infra/deploy/site/Containerfile -t jevidocs-site:local .
podman build -f infra/deploy/admin/Containerfile -t jevidocs-admin:local .
```
