---
title: Code blocks
description: Syntax-highlighted code with titles and copy buttons.
position: 7
---

Fenced code blocks are highlighted on the server with
[Chroma](https://github.com/alecthomas/chroma), which knows hundreds of
languages. The reader adds a copy button to every block.

## Usage

~~~mdx
```go title="main.go"
package main

func main() {
	println("hello, jevidocs")
}
```
~~~

```go title="main.go"
package main

func main() {
	println("hello, jevidocs")
}
```

## Title

`title="..."` after the language adds a caption with the file name. Without a
title the block has no caption:

```ts
const res = await fetch('https://api.jevidocs.jevido.app/api/projects')
const projects = await res.json()
```

## Languages

Use the usual names: `go`, `ts`, `js`, `svelte`, `sh`, `json`, `yaml`, `sql`,
`md`, `html`, `css`, `dockerfile`, … Unknown or missing languages render as
plain text.

```sql title="search.sql"
SELECT slug, title
FROM pages
WHERE to_tsvector('english', title || ' ' || plain) @@ to_tsquery('english', 'deploy:*');
```

```json
{ "type": "page", "name": "Introduction", "slug": "" }
```

<Callout type="info">
Highlighting uses CSS classes (`.chroma .k`, `.chroma .s`, …), styled with
GitHub's light and dark palettes by the reader.
</Callout>
