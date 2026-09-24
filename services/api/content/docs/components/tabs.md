---
title: Tabs
description: Show alternatives, like package managers or languages, one at a time.
position: 3
---

## Usage

~~~mdx
<Tabs items="bun,npm,pnpm">
<Tab value="bun">

```sh
bun add jevidocs
```

</Tab>
<Tab value="npm">npm install jevidocs</Tab>
<Tab value="pnpm">pnpm add jevidocs</Tab>
</Tabs>
~~~

<Tabs items="bun,npm,pnpm">
<Tab value="bun">

```sh
bun add jevidocs
```

</Tab>
<Tab value="npm">npm install jevidocs</Tab>
<Tab value="pnpm">pnpm add jevidocs</Tab>
</Tabs>

## Items

`items` lists the tab labels, either comma-separated or in fumadocs' JSX form:

```mdx
<Tabs items={['Go', 'TypeScript']}>
```

Without `items`, the labels are taken from each `<Tab value="...">` in order.
A `<Tab>` without `value` takes the next label from `items`. The first tab is
active.

<Tabs>
<Tab value="Go">

```go
fmt.Println("hello")
```

</Tab>
<Tab value="TypeScript">

```ts
console.log('hello')
```

</Tab>
</Tabs>
