# 0008. Analytics as daily aggregates, no personal data

- Status: accepted
- Date: 2026-09-24

## Context

Docs teams want to know which pages are read and what readers search for
without finding it. fumadocs leaves this to third-party analytics. We want it
built in without cookies, consent banners or storing who read what.

## Decision

- Page views and searches are counted per project, slug (or normalised
  query) and day with `INSERT ... ON CONFLICT DO UPDATE`. No IP addresses,
  user agents or sessions are stored; addresses only feed an in-memory rate
  limiter.
- The reader sends views with `sendBeacon` as `text/plain` so no CORS
  preflight is needed; known bots are ignored.
- The admin's Insights tab reads the aggregates, including zero-result
  searches as "content gaps".

## Consequences

- Cheap to store and query; nothing to anonymise or expire.
- No unique visitors, referrers or funnels.
- Counts from a single process's rate limiter are approximate under abuse.

## Alternatives considered

- **Raw event log:** richer, but personal data and unbounded growth.
- **External analytics (Plausible etc.):** another service and script on
  every page.
