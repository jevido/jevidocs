# 0011. A native OpenAPI reference, rendered by the site

- Status: accepted
- Date: 2026-09-25

## Context

The OpenAPI importer generated Markdown: an index, a page per tag and a page
per operation, with tables and static curl/JS/Go tabs. It fitted decision
0003 (the API renders, the site styles), but reads like a table dump. We
want the reference to feel like Scalar: the whole API on one page, two
columns, samples that follow the reader's chosen server, credentials and
language, a schema explorer with nested attributes, and a client to send
requests.

Those depend on state that only exists in the browser (which server, which
token, which language, which `oneOf` variant is open), so the HTML cannot be
produced once at save time.

## Decision

- A page gets a `kind`. An `openapi` page stores the OpenAPI document as its
  body. On save, the API parses it (`app/openapi`) into a `Reference`: tags,
  operations, parameters, bodies, responses, security schemes, servers and
  schemas, with descriptions rendered to HTML, `$ref`s to named schemas kept
  as references, `allOf` merged and examples generated. It is stored as JSON
  next to the page (`pages.api`), together with Markdown for agents
  (`pages.markdown`) and search sections per operation.
- The site renders the `Reference` with Svelte components and builds code
  samples, requests and the request client in the browser.
- An import makes one page (at the prefix) instead of a page per operation,
  and removes the old generated pages below it.
- Spec files take part in docs as code as `*.openapi.{json,yaml,yml}`.

## Consequences

- Parsing, `$ref` resolution and Markdown rendering still happen once, on
  save, in Go; reads stay one query. Search, `llms.txt`, `page.md` and MCP
  keep working through the derived sections and Markdown.
- The `Reference` JSON is a second contract between the API and the site
  (documented in `services/api/README.md`). Changing it means changing both,
  and bumping `docs.RenderVersion` so stored pages are re-derived.
- One page per API instead of one per operation: the sidebar shows the
  operations as an outline of that page, and search links to anchors
  (`#tag/pets/GET/pets`). Very large specs make a long page.
- Test Request is a browser `fetch`: it only works against APIs that allow
  the docs site's origin.

## Alternatives considered

- **Embed Scalar itself** (`@scalar/api-reference` from a CDN). Fastest to
  a Scalar look, but a large third-party bundle on every docs page, its own
  theming that fights ours, no search or llms.txt integration, and the spec
  parsed in every visitor's browser.
- **Keep generating Markdown, add more components.** Stays within 0003, but
  the reader's choices (server, auth, language, variants) cannot be baked
  into HTML, so samples and the client would still need client-side code and
  data; the data is the reference.
- **Parse the spec in the browser.** Keeps the API simpler, but ships a YAML
  parser and `$ref` resolver to readers and leaves search and agents without
  the content.
