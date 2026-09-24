# 0010. Private projects: members, share links, invisible otherwise

- Status: accepted
- Date: 2026-09-24

## Context

"Private" first meant "hidden from the docs site for everyone", which made
a private project unreadable even for its own admins. Teams need to share
internal docs with colleagues (who have accounts) and with clients (who do
not).

## Decision

- A private project is readable by admins, by users listed as its members,
  and by anyone presenting a valid share link. The check is one function
  (`store.ReadableProject`) used by every public read route and MCP read
  tool.
- Everyone else gets 404, not 403, so the existence of a private project is
  not revealed.
- The docs site signs readers in with the admin's accounts (bearer token in
  the reader's browser). Share tokens are stored hashed, can expire, record
  their last use, and are revoked by deleting them; the site keeps a token
  per project in localStorage and sends it as `X-Jevidocs-Share`.
- Responses for private projects are `Cache-Control: private, no-store`.

## Consequences

- Access is per project for members and links, while roles stay global:
  an editor who is not a member cannot read the project on the site, but
  can still edit it in the admin. Per-project admin permissions would be a
  new decision.
- Share links are bearer secrets: anyone with the URL can read.

## Alternatives considered

- **Signed, stateless share links** (like preview links): no table, but no
  revocation or last-used information.
- **Cookies on a shared parent domain:** would also work, but the site,
  admin and API already use bearer tokens and different hosts.
