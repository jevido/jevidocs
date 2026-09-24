---
title: llms.txt
description: Machine-readable versions of your documentation.
position: 32
section: Reference
---

Every project is available in forms made for language models, following
[llmstxt.org](https://llmstxt.org).

| URL | Content |
| --- | ------- |
| `/api/projects/{project}/llms.txt` | Title, summary and a link per page, in reading order |
| `/api/projects/{project}/llms-full.txt` | Every published page's Markdown, separated by `---` |
| `/api/projects/{project}/page.md?slug=a/b` | One page as Markdown |

For the jevidocs docs the reader site redirects the familiar addresses:

- [`https://jevidocs.jevido.app/llms.txt`](https://jevidocs.jevido.app/llms.txt)
- [`https://jevidocs.jevido.app/llms-full.txt`](https://jevidocs.jevido.app/llms-full.txt)
- `https://jevidocs.jevido.app/docs/<slug>.md`, e.g.
  [`/docs/mcp.md`](https://jevidocs.jevido.app/docs/mcp.md)

```md title="llms.txt"
# jevidocs

> Documentation framework on Go, Goravel, Svelte 5 and PostgreSQL …

## Docs

- [Introduction](https://api.jevidocs.jevido.app/api/projects/jevidocs/page.md?slug=): jevidocs is a documentation framework …
- [Quick start](https://api.jevidocs.jevido.app/api/projects/jevidocs/page.md?slug=quick-start): Run the API …
```

## In the reader

Every page has a **Copy Markdown** button and an **Open in** menu that starts
a ChatGPT or Claude conversation about the page, pointing it at the page's
`.md` URL.

<Callout type="idea">
Components stay as tags in the Markdown (`<Callout>`, `<Tabs>`), which models
read fine and which keeps the source faithful.
</Callout>
