# 0004. Content lives in Postgres; files sync into it

- Status: accepted
- Date: 2026-09-24

## Context

fumadocs reads content from the filesystem at build time. jevidocs wants an
admin and MCP tools that edit docs live, and still wants docs as code for
projects (including its own documentation) that keep Markdown in git.

## Decision

- Pages are rows in `pages`, keyed by project and slug. The slug is the path
  (`guides/install`); the page tree is derived from slugs, positions and
  sections (`app/docs/tree.go`).
- Files sync into a project: the API's own `content/docs` on every start
  (the `jevidocs` project is "managed"), and any folder through
  `PUT /api/admin/projects/{p}/sync` (the `jevidocs push` CLI).

## Consequences

- Admin/MCP edits to a managed project are overwritten on the next start;
  the files win.
- One source of truth per project, chosen by how it is edited.
- Migrations must stay compatible with the previous release during rolling
  deploys.

## Alternatives considered

- **Git as the only store** (commit on every admin save): heavy for an admin
  and MCP write path.
- **meta.json files for ordering:** replaced by `position` and `section`
  front matter, which also work for database-only pages.
