---
title: Versions
description: Keep documentation for several releases side by side.
position: 6
---

Like fumadocs, jevidocs versions documentation by keeping **one project per
version** and linking them. Each version has its own pages, tree, search and
`llms.txt`; the reader adds a switcher to the navbar.

<Steps>
<Step>

### Create a project per version

For example `acme` for the current release and `acme-v1` for the previous
one. Copy pages across with the CLI:

```sh title="terminal"
jevidocs pull acme-docs -project acme
jevidocs push acme-docs -project acme-v1
```

</Step>
<Step>

### Give them the same version group

In each project's **Settings**, set **Version** to the same group (for
example `acme`) and a label per project (`v2`, `v1`). Over the API these are
`version_group` and `version_label` on the project.

</Step>
<Step>

### Switch in the reader

Public projects that share a group appear in a dropdown next to the project
name, newest label first. Labels sort naturally, so `v10` comes before `v9`
and `1.10` before `1.9`. Switching keeps the current page slug; if that page
does not exist in the other version, the reader shows its not-found page.

</Step>
</Steps>

The public project endpoint lists the group:

```json title="GET /api/projects/acme"
{
  "slug": "acme",
  "version_group": "acme",
  "version_label": "v2",
  "versions": [
    { "slug": "acme", "name": "Acme", "label": "v2", "url": "https://jevidocs.jevido.app/p/acme" },
    { "slug": "acme-v1", "name": "Acme", "label": "v1", "url": "https://jevidocs.jevido.app/p/acme-v1" }
  ]
}
```

<Callout type="info">
`versions` is only present when at least two public projects share the
group.
</Callout>
