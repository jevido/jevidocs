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

### History

Every save that changes a page's title, description or body keeps the
previous version. **History** in the editor lists the last 50 versions. Pick
one to see its Markdown or a line diff against what is in the editor, then
**Restore this version** to bring it back. Restoring is a save too, so the
version it replaces stays in the history and nothing is lost.

## Assets

The **Assets** tab of a project stores images and files (PNG, JPEG, GIF,
WebP, SVG, AVIF, PDF and plain text, up to 8 MB each) in the database. Drop
files on it or choose them, then **Copy Markdown** to get `![name](url)`.

In the page editor you can also paste or drop an image straight into the
Markdown: it is uploaded and the `![name](url)` is inserted where the cursor
was. Uploading the same bytes twice reuses the existing asset. Asset URLs
never change, so they are cached for a year.

## Insights

The **Insights** tab of a project shows the last 7, 30 or 90 days:

- **Page views** per day, counted once per page load in the docs reader
  (bots and unknown pages are ignored; nothing identifies the reader).
- **Top pages** by views, linking to the editor.
- **Top searches** from the search dialog, with how many results each found.
- **Content gaps**: searches that found nothing. These are the pages worth
  writing next.

The dashboard shows the total views of every project over 30 days.

## Feedback

Every docs page ends with "Was this page helpful?". Readers answer with a
thumbs up or down and can add a message. The **Feedback** tab of a project
shows a helpful score per page and every answer, newest first; tick "Only
with a message" to read just the comments. The API accepts at most 20
answers per IP address per hour.

## Users

**Users** lists everyone who can sign in to the admin; each of them can edit
every project. Add a user with a name, email and a password of at least 8
characters, or delete one (their sessions and API tokens stop working). You
cannot delete yourself. The same page changes your own password.

## API tokens

**Tokens** creates long-lived API tokens for scripts and the
[MCP server](/docs/mcp). A token is shown once, together with a ready-to-paste
`.mcp.json` snippet; only its SHA-256 hash is stored. Revoke tokens you no
longer use.
