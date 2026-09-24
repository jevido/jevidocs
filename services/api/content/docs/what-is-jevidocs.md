---
title: What is jevidocs?
description: How jevidocs relates to fumadocs, and what each piece maps to.
position: 2
section: Get started
---

[fumadocs](https://fumadocs.dev) is a React / Next.js documentation framework:
content lives in MDX files, is compiled at build time, and is rendered by React
components. jevidocs keeps the reader experience and the content conventions,
but moves the content into a database behind an API, so docs can be edited in
a browser or by an agent without a rebuild.

## What maps to what

| fumadocs | jevidocs |
| -------- | -------- |
| `fumadocs-core` source loader and page tree | Go package `services/api/app/docs` (`BuildTree`, breadcrumbs, neighbours) |
| `fumadocs-mdx` (MDX files, compiled at build) | Markdown in PostgreSQL, rendered to HTML on save |
| MDX components (`<Callout>`, `<Tabs>`, …) | The same tags, expanded server-side to HTML |
| Shiki | Chroma, with GitHub light and dark colours |
| `fumadocs-ui` `DocsLayout` / `DocsPage` | `apps/site`, a Svelte 5 docs reader |
| Orama search | PostgreSQL full-text search (`to_tsvector`, `ts_headline`) |
| `meta.json` ordering and separators | `position` and `section` front matter |
| `llms.txt` / `.mdx` routes | `/api/projects/{p}/llms.txt`, `llms-full.txt`, `page.md` |
| Content in git | Content in Postgres, edited in the admin app or over MCP |
| — | REST API, admin app, API tokens, MCP server |

## Differences worth knowing

<Accordions>
<Accordion title="No JSX runtime">
Components are written like JSX but there is no JavaScript evaluation.
Attributes are plain strings (`items={['a', 'b']}` is understood for tabs), and
component tags must sit on their own lines. See
[Markdown](/docs/writing/markdown).
</Accordion>
<Accordion title="Rendered once, served many times">
A page's HTML, table of contents and search sections are computed when the
page is saved and stored next to it. Reading a page is a single row lookup.
</Accordion>
<Accordion title="A single-page reader">
The reader is a Svelte app that fetches pages from the API, not a statically
generated site. Machine readers should use `llms.txt` and the `.md` URLs.
</Accordion>
<Accordion title="Many projects">
One deployment serves many documentation projects, each with its own tree,
navbar links and search.
</Accordion>
</Accordions>
