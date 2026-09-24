---
title: Page tree
description: How slugs become the sidebar.
position: 3
---

The sidebar is a **page tree**, built by `docs.BuildTree` in the API, with the
same shape as fumadocs' `PageTree`: pages, folders and separators.

## Rules

<Steps>
<Step>

### A slug is a path

`guides/install` lives in the folder `guides`. The empty slug `""` is the
project's index page. Files named `index.md` map to their folder's slug.

</Step>
<Step>

### Folders come from prefixes

Any slug prefix with children is a folder. A page whose slug equals the folder
path (`guides`) is the folder's **index**: it names the folder, gives it its
icon, and opens when the folder title is clicked. A folder without an index is
named after its path segment (`api-reference` → "Api reference").

</Step>
<Step>

### Siblings sort by position

Siblings sort by `position`, then by title. A folder sorts by its index page's
position, or by its lowest child's when it has no index.

</Step>
<Step>

### Sections make separators

A top-level page or folder whose `section` differs from the one before it
starts a separator with that label.

</Step>
</Steps>

## Example

<Files>
<Folder name="content/docs" defaultOpen>
<File name='index.md — slug "", section Get started' />
<File name="quick-start.md — position 1" />
<Folder name="writing (index.md, section Writing)" defaultOpen>
<File name="markdown.md" />
<File name="page-tree.md" />
</Folder>
</Folder>
</Files>

## Reading order

The tree also defines reading order: a folder's index, then its children, top
to bottom. That order drives the previous/next links under every page,
breadcrumbs, and the order of `llms.txt`.

```json title="GET /api/projects/jevidocs → tree"
{
  "name": "jevidocs",
  "children": [
    { "type": "page", "name": "Introduction", "slug": "" },
    { "type": "separator", "name": "Writing" },
    {
      "type": "folder", "name": "Writing",
      "index": { "type": "page", "name": "Writing", "slug": "writing" },
      "children": [{ "type": "page", "name": "Markdown", "slug": "writing/markdown" }]
    }
  ]
}
```
