# 0005. Hosted MCP lives in the API

- Status: accepted
- Date: 2026-09-24

## Context

Agents should read, search and edit documentation. The operations are the
same as the REST API's, with the same validation and permissions.

## Decision

- The MCP server is part of `services/api` (`app/mcpserver`), mounted at
  `/mcp` with the official Go SDK over streamable HTTP, stateless, JSON
  responses (as in jevido/work's 0005).
- Tools are thin wrappers over `app/store`. Read tools are public; write
  tools need an admin API token as a Bearer header.

## Consequences

- A tool and its REST route cannot drift apart.
- Tool calls must finish within `http.request_timeout` (15s).
- Recursive results (page trees) go out without an output schema, since the
  SDK cannot derive one for recursive types.

## Alternatives considered

- **A separate MCP service:** duplicates auth and data access.
- **Local stdio MCP in the CLI:** useful later, but the hosted one works
  with any client without installing anything.
