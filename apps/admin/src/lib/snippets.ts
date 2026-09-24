// Component snippets for the editor toolbar, in the syntax the API renders
// (see services/api/README.md, "Content").
export const snippets: { label: string; text: string }[] = [
  {
    label: 'Callout',
    text: '<Callout type="info" title="Note">\nSomething worth knowing.\n</Callout>\n',
  },
  {
    label: 'Cards',
    text:
      '<Cards>\n<Card title="Getting started" href="/docs" description="Install and write your first page" />\n<Card title="Components" href="/docs/components" description="Everything Markdown can do" />\n</Cards>\n',
  },
  {
    label: 'Tabs',
    text: '<Tabs items="bun,npm">\n<Tab value="bun">\n\n```sh\nbun install\n```\n\n</Tab>\n<Tab value="npm">\n\n```sh\nnpm install\n```\n\n</Tab>\n</Tabs>\n',
  },
  {
    label: 'Steps',
    text: '<Steps>\n<Step>\n\n### Install\n\nGet the tools.\n\n</Step>\n<Step>\n\n### Write\n\nAdd your first page.\n\n</Step>\n</Steps>\n',
  },
  {
    label: 'Accordion',
    text: '<Accordions>\n<Accordion title="Why would I?">\nBecause it is nice.\n</Accordion>\n</Accordions>\n',
  },
  {
    label: 'Files',
    text:
      '<Files>\n<Folder name="app" defaultOpen>\n<File name="main.go" />\n<File name="go.mod" />\n</Folder>\n<File name="README.md" />\n</Files>\n',
  },
  {
    label: 'Code',
    text: '```go title="main.go"\npackage main\n\nfunc main() {}\n```\n',
  },
]
