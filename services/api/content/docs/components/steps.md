---
title: Steps
description: Numbered instructions with a line connecting them.
position: 4
---

## Usage

Wrap each step in `<Step>`. Start a step with a heading; leave blank lines
around the Markdown inside.

```mdx
<Steps>
<Step>

### Install

Run `bun install`.

</Step>
<Step>

### Configure

Copy `.env.example` to `.env`.

</Step>
</Steps>
```

<Steps>
<Step>

### Install

Run `bun install`.

</Step>
<Step>

### Configure

Copy `.env.example` to `.env`.

</Step>
<Step>

### Start

Run `go run .` and open the reader.

</Step>
</Steps>

<Callout type="info">
Step headings are ordinary headings, so they appear in the table of contents.
</Callout>
