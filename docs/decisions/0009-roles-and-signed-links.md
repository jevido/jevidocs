# 0009. Three roles and HMAC-signed links

- Status: accepted
- Date: 2026-09-24

## Context

More than one person edits docs, not all of them should manage users or
delete projects. Drafts need to be shown to reviewers without an account,
and images of private projects must not be public.

## Decision

- Users are `admin`, `editor` or `viewer`. One middleware decides from the
  HTTP method and path, so new admin routes are editor-only for writes by
  default; project and user management is admin-only.
- Draft previews and private asset URLs are signed with HMAC-SHA256 over
  the page or asset id (and an expiry for previews) using `APP_KEY`. No
  table of links.

## Consequences

- Stateless: nothing to store or clean up. Rotating `APP_KEY` invalidates
  every preview link and private asset URL at once.
- A preview link cannot be revoked before it expires except by rotating the
  key.
- Roles are global, not per project.

## Alternatives considered

- **Per-project memberships:** finer control, but a data model and UI
  nobody needed yet.
- **Stored share tokens:** revocable, but another table and cleanup job.
