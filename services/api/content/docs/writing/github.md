---
title: Sync from GitHub
description: Keep a project's pages in a GitHub repository and sync them on every push.
position: 5
---

A project can take its pages from a folder of Markdown in a GitHub
repository: docs as code, without running the [CLI](/docs/quick-start) in CI.
The API downloads the repository, reads every `.md` and `.mdx` file under the
folder and syncs them into the project, the same way the API's own
`content/docs` is synced on start.

<Steps>
<Step>

### Connect the repository

In the admin, open the project, go to **Settings → GitHub source** and fill
in the repository (`owner/name`), the branch or tag (default `main`) and the
folder (default `docs`; empty for the whole repository). Save.

</Step>
<Step>

### Sync

Press **Sync now**. The sync runs in the background; the status line shows
when it finished and how many pages were created, updated, left unchanged or
deleted, or the error.

</Step>
<Step>

### Sync on push (optional)

Add a webhook in the repository under **Settings → Webhooks** with the
payload URL, `application/json` content type and secret shown in the admin,
for the push event only. Every push to the configured branch starts a sync.

</Step>
</Steps>

<Callout type="warn" title="The repository wins">
Pages that have no file in the folder are deleted on every sync, and edits
made in the admin are overwritten by the next one. Change the files instead.
</Callout>

## How files become pages

The rules are those of the [page tree](/docs/writing/page-tree): the path
below the folder is the slug, `index.md` is a folder's page, and front matter
sets title, description, position, section and `root`.

Repositories written for other tools mostly just work:

- Route-group folders such as `(framework)/` (Next.js, fumadocs) are left out
  of the slug: `(framework)/comparisons.mdx` becomes `comparisons`.
- Segments are lowercased and other characters become dashes:
  `UI/Theme Options.md` becomes `ui/theme-options`.
- MDX is read as Markdown with jevidocs' [components](/docs/components);
  imports and custom JSX components show up as text.

## Limits

| Limit | Value |
| ----- | ----- |
| Download size | 50 MB (gzip) |
| Download time | 20 seconds |
| Files | 2000 Markdown files |
| File size | 1 MB each (larger files are skipped) |

Public repositories need nothing else. For private ones, set `GITHUB_TOKEN`
on the API; it is sent when downloading the archive.

## API

| Method | Path | Body → Returns |
| ------ | ---- | -------------- |
| GET | `/api/admin/projects/{project}/source` | → `{repo, ref, path, secret, synced_at, status, webhook_url}` |
| PUT | `/api/admin/projects/{project}/source` | `{repo, ref, path}` → same (an empty `repo` disconnects) |
| POST | `/api/admin/projects/{project}/source/sync` | → `202 {started}`; poll `GET .../source` |
| POST | `/api/hooks/github/{project}` | GitHub push event, signed with `X-Hub-Signature-256` |

The webhook answers `401` for a bad signature, `{pong: true}` to GitHub's
ping, ignores pushes to other branches, and answers `202` when it starts a
sync.
