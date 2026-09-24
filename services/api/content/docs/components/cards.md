---
title: Cards
description: A responsive grid of linked cards.
position: 2
---

## Usage

```mdx
<Cards>
<Card title="Quick start" href="/docs/quick-start" description="Run it locally." />
<Card title="MCP" href="/docs/mcp">
Connect an agent to your docs.
</Card>
</Cards>
```

<Cards>
<Card title="Quick start" href="/docs/quick-start" description="Run it locally." />
<Card title="MCP" href="/docs/mcp">
Connect an agent to your docs.
</Card>
</Cards>

## Card attributes

| Attribute | Meaning |
| --------- | ------- |
| `title` | Card heading. |
| `href` | Link target. Without it the card is not a link. |
| `description` | Text under the title. The card's body is used when omitted. |

A card without `href`:

<Cards>
<Card title="Just information" description="Not every card needs to go somewhere." />
</Cards>
