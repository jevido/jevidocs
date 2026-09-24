---
title: Files
description: Display a file and folder structure.
position: 6
---

## Usage

```mdx
<Files>
<Folder name="services" defaultOpen>
<Folder name="api" defaultOpen>
<File name="main.go" />
<File name="go.mod" />
</Folder>
</Folder>
<File name="package.json" />
</Files>
```

<Files>
<Folder name="services" defaultOpen>
<Folder name="api" defaultOpen>
<File name="main.go" />
<File name="go.mod" />
</Folder>
</Folder>
<File name="package.json" />
</Files>

## Attributes

| Element | Attribute | Meaning |
| ------- | --------- | ------- |
| `Folder` | `name` | Folder label. |
| `Folder` | `defaultOpen` | Start expanded. |
| `File` | `name` | File label. |

Folders are `<details>` elements, so they toggle without JavaScript.
