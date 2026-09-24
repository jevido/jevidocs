---
title: Callout
description: Draw attention to a note, a warning or a tip.
position: 1
---

## Usage

```mdx
<Callout type="warn" title="Heads up">
This changes **production** data.
</Callout>
```

<Callout type="warn" title="Heads up">
This changes **production** data.
</Callout>

## Types

`type` is one of `info` (default, also `note`), `warn` (`warning`), `error`
(`danger`), `success` and `idea` (`tip`). `title` is optional.

<Callout>A plain `info` callout without a title.</Callout>

<Callout type="error" title="Error">
Something went wrong.
</Callout>

<Callout type="success" title="Success">
Everything worked.
</Callout>

<Callout type="idea" title="Idea">
Try linking to a [related page](/docs/components/cards).
</Callout>

## Attributes

| Attribute | Values | Default |
| --------- | ------ | ------- |
| `type` | `info`, `warn`, `error`, `success`, `idea` | `info` |
| `title` | any text | none |
