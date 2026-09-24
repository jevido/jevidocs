---
title: Front matter
description: Page metadata at the top of a Markdown document.
position: 2
---

A page can start with a `---` block of `key: value` lines. Values may be
quoted. Only the keys below are read; everything else of YAML is ignored.

```md title="guides/install.md"
---
title: Installation
description: Get jevidocs running.
icon: download
position: 1
section: Guides
---

## Requirements
```

| Key | Meaning |
| --- | ------- |
| `title` | Page title; shown as the heading, in the sidebar and in search. |
| `description` | Shown under the title, in `llms.txt` and in search. |
| `icon` | Optional sidebar icon name. |
| `position` | Integer sort order among siblings (`order` works too). |
| `section` | For top-level pages: starts a sidebar separator with this label. |

<Callout type="info" title="Fields win over front matter">
When a page is saved through the admin app or the API, explicit fields
(title, description, …) take precedence; front matter fills in whatever is
left empty. The front matter block itself is removed from the stored body.
</Callout>
