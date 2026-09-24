# 0003. Render Markdown and components in the API

- Status: accepted
- Date: 2026-09-24

## Context

fumadocs authors write MDX with JSX components (`<Callout>`, `<Tabs>`, ...)
compiled by the bundler. Our pages live in a database and are edited at
runtime, so there is no build step, and three consumers (site, admin preview,
MCP/llms.txt) need the same result.

## Decision

- The API renders each page when it is saved: Markdown to HTML, a table of
  contents, search sections and plain text are stored with the page.
- Components are written JSX-style on their own lines and expanded to plain
  HTML with `fd-*` classes before Markdown parsing, so Markdown inside them
  still renders. The HTML contract is documented in `services/api/README.md`.
- Clients only style that HTML and attach behaviour (tab switching, copy
  buttons).

## Consequences

- Reading a page is one indexed query; no rendering on the read path.
- Adding a component means changing the expander and both stylesheets.
- Only the listed components exist; arbitrary JSX is not supported.
- Raw HTML in pages is allowed: authors are admins with tokens, the same
  trust as the site itself.

## Alternatives considered

- **Render in the browser** (markdown-it + Svelte components): every client
  would need the renderer, and MCP/llms.txt would get different output.
- **A real MDX pipeline in Node:** a second server runtime for one feature.
