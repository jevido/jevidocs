---
title: Theming
description: Give a project its own accent colour and logo, and learn the reader's keyboard shortcuts.
position: 4
---

Every project can change how the reader looks without touching code. Open
the project in the [admin](/docs/admin), go to **Settings**, and set:

| Field | Meaning |
| ----- | ------- |
| **Accent** | A preset name, or a CSS colour: `#rgb`, `#rrggbb`, `hsl(262 83% 58%)` or `oklch(0.6 0.2 250)`. Empty uses the default purple. |
| **Logo URL** | An `https://` image shown next to the project name in the navbar, instead of the jevidocs mark. |

The accent colours links, the active sidebar item, the table of contents
highlight, focus rings and the loading bar.

## Presets

Like fumadocs' colour presets, five names pick a tuned pair of light and dark
colours:

<Cards>
<Card title="neutral" description="Greys only, for a quiet site." />
<Card title="ocean" description="A calm blue." />
<Card title="purple" description="The jevidocs default, a bit more vivid." />
<Card title="emerald" description="A fresh green." />
<Card title="ruby" description="A warm red." />
</Cards>

<Callout type="info" title="Why so strict?">
The accent ends up in CSS, so the API accepts only presets and the colour
formats above. Anything else, like `red` or `var(--x)`, is rejected with a
validation error.
</Callout>

## Over the API

`accent` and `logo_url` are ordinary project fields:

```sh title="terminal"
curl -X PUT https://api.jevidocs.jevido.app/api/admin/projects/demo \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"Demo","accent":"ocean","logo_url":"https://example.com/logo.svg"}'
```

## Reading shortcuts

| Keys | Action |
| ---- | ------ |
| <kbd>⌘</kbd> <kbd>K</kbd> / <kbd>Ctrl</kbd> <kbd>K</kbd> | Open search |
| <kbd>[</kbd> or <kbd>Alt</kbd> <kbd>←</kbd> | Previous page |
| <kbd>]</kbd> or <kbd>Alt</kbd> <kbd>→</kbd> | Next page |
| <kbd>Esc</kbd> | Close search or menus |

The sidebar can be collapsed from the button at its bottom on wide screens;
the choice is remembered in the browser. On narrow screens the table of
contents moves into an **On this page** dropdown above the article, and
printing a page leaves out the navigation.
