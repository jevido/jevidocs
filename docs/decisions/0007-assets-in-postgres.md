# 0007. Assets are stored in Postgres

- Status: accepted
- Date: 2026-09-24

## Context

Docs need images. The API runs as a rolling-deployed container with no
persistent volume, on one VPS, next to a Postgres resource that is already
backed up daily. There is no object store in the stack.

## Decision

- Uploaded files live in the `assets` table as `bytea`, deduplicated per
  project by SHA-256, at most 8 MB each, from an allow-list of image, PDF and
  text types.
- They are served by the API at `/api/assets/{id}/{name}` with an immutable
  one-year cache, `nosniff`, and a locked-down CSP for SVG.

## Consequences

- Assets survive redeploys and are in the same backups as the pages that
  reference them; no second storage system to run.
- The database grows with every upload and backups get larger; a docs site
  with many large images would outgrow this.
- Every asset request hits the API (browsers and proxies cache it after the
  first load).

## Alternatives considered

- **A volume on the API container:** breaks rolling deploys with two
  containers and is not backed up.
- **S3-compatible object storage:** the right answer at scale, but a new
  service and credentials for a first version.
