---
title: Components
description: The building blocks you can use inside a page.
position: 20
---

jevidocs understands the most used fumadocs UI components. They are written
with JSX-style tags on their own lines and expanded to HTML on the server, so
every client (the reader, the admin preview, `llms-full.txt` consumers) sees
the same result.

<Cards>
<Card title="Callout" href="/docs/components/callout" description="Info, warning, error, success and idea boxes." />
<Card title="Cards" href="/docs/components/cards" description="A grid of linked cards." />
<Card title="Tabs" href="/docs/components/tabs" description="Switch between alternatives." />
<Card title="Steps" href="/docs/components/steps" description="Numbered, step-by-step instructions." />
<Card title="Accordion" href="/docs/components/accordion" description="Collapsible questions and details." />
<Card title="Files" href="/docs/components/files" description="A file tree." />
<Card title="Code blocks" href="/docs/components/code-blocks" description="Highlighted code with titles and copy buttons." />
</Cards>

## Rendered HTML

The server emits final HTML with `fd-*` classes, for example
`<div class="fd-callout" data-type="warn">`. A client only has to style those
classes and attach behaviour (tab switching, copy buttons). The full contract
is in `services/api/README.md`.

<Callout type="idea" title="Writing your own reader">
Because the HTML is final, any frontend can render jevidocs pages: fetch
`/api/projects/{project}/page?slug=...` and style the `fd-*` classes.
</Callout>
