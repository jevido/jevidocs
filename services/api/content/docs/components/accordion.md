---
title: Accordion
description: Collapsible sections for FAQs and details.
position: 5
---

## Usage

```mdx
<Accordions>
<Accordion title="Is it free?">
Yes, jevidocs is open source.
</Accordion>
<Accordion title="Does it need Node.js?" id="node">
Only to build the reader and admin; the API is a single Go binary.
</Accordion>
</Accordions>
```

<Accordions>
<Accordion title="Is it free?">
Yes, jevidocs is open source.
</Accordion>
<Accordion title="Does it need Node.js?" id="node">
Only to build the reader and admin; the API is a single Go binary.
</Accordion>
</Accordions>

## Attributes

| Attribute | Meaning |
| --------- | ------- |
| `title` | The clickable summary. |
| `id` | Optional element id, so you can link to `#node`. |

Accordions are native `<details>` elements: they work without JavaScript and
are searchable with the browser's find.
