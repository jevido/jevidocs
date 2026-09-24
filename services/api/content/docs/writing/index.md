---
title: Writing
description: How pages are written, organised and rendered.
position: 10
section: Writing
---

Every page is a Markdown document with a small front matter block. Pages are
stored per project and identified by their **slug**, a path like
`guides/install`. The slugs alone decide the folder structure of the sidebar.

<Cards>
<Card title="Markdown" href="/docs/writing/markdown" description="CommonMark, GFM and JSX-style components." />
<Card title="Front matter" href="/docs/writing/front-matter" description="Title, description, icon, position and section." />
<Card title="Page tree" href="/docs/writing/page-tree" description="How slugs become folders, order and separators." />
<Card title="Components" href="/docs/components" description="Everything you can put inside a page." />
</Cards>

## Where pages come from

<Tabs items="Admin app,MCP,Files">
<Tab value="Admin app">

Create and edit pages in the [admin app](/docs/admin) with a live preview.

</Tab>
<Tab value="MCP">

Let an agent write them with the `create_page` and `update_page` tools of the
[MCP server](/docs/mcp).

</Tab>
<Tab value="Files">

The `jevidocs` project itself is synced from `services/api/content/docs` on
every API start. Files are the source of truth for that project: pages without
a file are removed.

</Tab>
</Tabs>
