---
title: Admin app
description: Manage projects, pages and API tokens in the browser.
position: 25
---

The admin app lives at
[admin.jevidocs.jevido.app](https://admin.jevidocs.jevido.app) (locally on port
4740). It is a Svelte 5 single-page app that talks to the [admin
endpoints](/docs/api#admin) of the API.

## Signing in

Sign in with an admin email and password. The first admin is created by the
API on start from `ADMIN_EMAIL` and `ADMIN_PASSWORD` when no user exists. A
successful login returns a session token that the app keeps in
`localStorage` and sends as `Authorization: Bearer <token>`; signing out
revokes it.

## Projects

<Steps>
<Step>

### Create a project

Open **Projects**, choose **New project** and fill in a slug (lowercase letters,
digits and dashes), a name and a description.

</Step>
<Step>

### Configure the navbar

Add a GitHub URL for the navbar icon and any number of navbar links
(`text` + `url`). Untick **Public** to hide the project from the reader and
the public API.

</Step>
<Step>

### Read it

Public projects appear at `https://jevidocs.jevido.app/p/<slug>`. The
dashboard links to each project's docs.

</Step>
</Steps>

## Editing pages

A project's page list is shown in tree order, indented by slug depth, with
its published state, section and position. The page editor has:

- fields for slug, title, description, icon, position, section and published;
- a Markdown editor (<kbd>Tab</kbd> indents, <kbd>Ctrl</kbd>/<kbd>⌘</kbd> +
  <kbd>S</kbd> saves);
- a toolbar that inserts [components](/docs/components);
- a live preview, rendered by the API's `POST /api/admin/preview` so it is
  exactly what the reader will show;
- write, split and preview modes and an unsaved-changes warning.

<Callout type="warn" title="The jevidocs project is managed">
The `jevidocs` project is synced from the repository on every API start.
Edits made to it in the admin app are overwritten on the next deploy; change
the files in `services/api/content/docs` instead.
</Callout>

## API tokens

**Tokens** creates long-lived API tokens for scripts and the
[MCP server](/docs/mcp). A token is shown once, together with a ready-to-paste
`.mcp.json` snippet; only its SHA-256 hash is stored. Revoke tokens you no
longer use.
