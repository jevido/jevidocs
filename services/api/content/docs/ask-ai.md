---
title: Ask AI
description: Let readers ask questions and get answers written from your docs.
position: 36
section: Reference
---

With Ask AI on, the search dialog (⌘K) gets an **Ask AI** row. The reader's
question is answered by Claude from your documentation only, with links to
the pages it used.

## How it works

<Steps>
<Step>

### Retrieve

The API runs the normal [search](/docs/search) for the question and takes
up to six matching sections (a heading's text, or the start of a page).

</Step>
<Step>

### Answer

It sends those excerpts and the question to the Claude Messages API with a
system prompt that restricts the answer to the excerpts and asks for
citations. The answer is short Markdown.

</Step>
<Step>

### Show

The dialog renders the answer (HTML is escaped; only paragraphs, lists,
code, bold and links are kept) and lists the source pages.

</Step>
</Steps>

## Enable it

Set an Anthropic API key on the API:

```sh title=".env"
ANTHROPIC_API_KEY=sk-ant-...
# optional
ANTHROPIC_MODEL=claude-opus-5-5
```

Without a key the feature is off: the project view reports `ask: false`, the
dialog shows no Ask row, and the endpoint answers 404.

## Endpoint

```http
POST /api/projects/{project}/ask
Content-Type: application/json

{"question": "How do I deploy on Coolify?"}
```

```json
{
  "answer": "Create three applications ... see [Coolify](https://jevidocs.jevido.app/docs/self-hosting/coolify).",
  "sources": [{ "title": "Coolify", "url": "https://jevidocs.jevido.app/docs/self-hosting/coolify" }]
}
```

<Callout type="warn" title="Limits">
Questions are 3 to 500 characters, and each client address may ask 20 per
hour. A call must finish within 12 seconds, so the model runs at low effort.
</Callout>
