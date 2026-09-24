# 0002. Go, Goravel, Svelte 5 and PostgreSQL

- Status: accepted
- Date: 2026-09-24

## Context

fumadocs is React/Next.js with MDX compiled at build time. The goal is the
same product on Go + Svelte 5 + PostgreSQL, with an admin, a REST API and an
MCP server, and no SvelteKit.

## Decision

- **API:** Go 1.27 with Goravel (routing, config, ORM, migrations, hashing).
- **Markdown:** goldmark (CommonMark + GFM + footnotes) with Chroma for
  highlighting.
- **UIs:** Svelte 5 (runes) + Vite + TypeScript as static single-page apps
  served by nginx.
- **Data:** PostgreSQL 18, full-text search with `to_tsvector`/`to_tsquery`.

## Consequences

- One Go binary does rendering, search, llms.txt and MCP; the UIs are static.
- No MDX runtime: components are a fixed set expanded by the API (see 0003).
- Search quality depends on Postgres' English stemmer; prefix matching and
  an ILIKE fallback cover partial words and code identifiers.

## Alternatives considered

- **SvelteKit with server rendering:** better SEO, but ruled out by the
  stack rules inherited from jevido/work.
- **Orama/Meilisearch:** better ranking, but another service to run.
