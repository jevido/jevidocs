---
title: MCP server
description: Let Claude and other agents read, search and write your documentation.
position: 31
section: Reference
---

The API hosts a [Model Context Protocol](https://modelcontextprotocol.io)
server at `https://api.jevidocs.jevido.app/mcp`, built with the official Go
SDK. It uses the **streamable HTTP** transport, **stateless**, answering with
plain JSON.

## Connect

<Tabs items="Claude Code,.mcp.json">
<Tab value="Claude Code">

Read-only access:

```sh
claude mcp add --transport http jevidocs https://api.jevidocs.jevido.app/mcp
```

With write access, pass an [API token](/docs/api/authentication):

```sh
claude mcp add --transport http jevidocs https://api.jevidocs.jevido.app/mcp \
  --header "Authorization: Bearer jd_…"
```

</Tab>
<Tab value=".mcp.json">

```json title=".mcp.json"
{
  "mcpServers": {
    "jevidocs": {
      "type": "http",
      "url": "https://api.jevidocs.jevido.app/mcp",
      "headers": {
        "Authorization": "Bearer jd_…"
      }
    }
  }
}
```

</Tab>
</Tabs>

## Tools

| Tool | Arguments | Needs token |
| ---- | --------- | ----------- |
| `ping` | — | no |
| `list_projects` | — | no |
| `get_page_tree` | `project` | no |
| `read_page` | `project`, `slug` | no |
| `search_docs` | `project`, `query` | no |
| `create_page` | `project`, `slug`, `title`, `description?`, `icon?`, `position?`, `section?`, `body`, `published?` | yes |
| `update_page` | `project`, `slug`, and any field to change | yes |
| `delete_page` | `project`, `slug` | yes |

- `read_page` returns `{title, description, url, markdown}`.
- `search_docs` returns up to 15 page and heading results with snippets.
- `update_page` keeps every field you omit. The slug identifies the page and
  cannot be changed through MCP.
- Bodies use the same Markdown and [components](/docs/components) as the
  admin app, including front matter.

<Callout type="info" title="Same rules as the REST API">
Every tool calls the same application code as the matching REST route, so
validation (unique slugs, required titles) is identical.
</Callout>

## Example prompts

- "Search the jevidocs docs for how search works and summarise it."
- "Read `quick-start` in project `demo` and add a Troubleshooting section."
- "Create a page `guides/deploy` in project `demo` from this README."

## Limits

Tool calls must finish within the API's request timeout (a few seconds):
Goravel's global timeout middleware buffers responses, which is also why the
transport answers with JSON rather than an event stream. There are no
server-initiated notifications.
