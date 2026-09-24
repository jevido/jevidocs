---
title: Search
description: Full-text search over pages and headings, in PostgreSQL.
position: 33
section: Reference
---

Search runs in PostgreSQL; there is no separate search service or index to
build. Open it in the reader with <kbd>⌘</kbd> <kbd>K</kbd> or
<kbd>Ctrl</kbd> <kbd>K</kbd>, or call the API:

```sh
curl 'https://api.jevidocs.jevido.app/api/projects/jevidocs/search?q=deploy'
```

```json title="Result"
[
  {
    "type": "heading",
    "slug": "self-hosting/coolify",
    "hash": "watch-paths",
    "title": "Watch paths",
    "page_title": "Coolify",
    "snippet": "A push to main triggers a <mark>deploy</mark> of …",
    "url": "https://jevidocs.jevido.app/docs/self-hosting/coolify#watch-paths"
  }
]
```

## How it works

<Steps>
<Step>

### Indexing on save

When a page is saved, its rendered HTML is split at `h2`–`h4` headings into
plain-text sections, stored as JSON next to the page, together with the page's
full plain text. A GIN index covers
`to_tsvector('english', title || ' ' || description || ' ' || plain)`.

</Step>
<Step>

### Matching pages

The query is split into words; each becomes a prefix term (`deploy:*`) and
all must match. Pages are ranked with `ts_rank`, and `ts_headline` builds the
snippet. When nothing matches, a plain substring match on title and text is
the fallback, for names the English stemmer mangles.

</Step>
<Step>

### Matching headings

For each matching page, up to three sections whose heading or text contains a
query word are returned as `heading` results, linking to `page#heading`.

</Step>
<Step>

### Similar results

When neither full-text nor substring search finds anything, typos get a
second chance: page titles and section headings are compared with the query
by trigram similarity (`pg_trgm`). Only hits close to the best score are
kept, and each is marked `"fuzzy": true`; the search dialog then says
"Showing similar results". Searching `calout` finds **Callout**.

The migration enables `pg_trgm` when the Postgres server offers it and skips
it otherwise; without the extension this step is simply left out.

</Step>
</Steps>

<Callout type="info">
Snippets are HTML-escaped on the server; only `<mark>` tags are added.
</Callout>
