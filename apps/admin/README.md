# admin

The jevidocs admin app (https://admin.jevidocs.jevido.app): Svelte 5 + Vite,
no SvelteKit. Sign in, manage documentation projects and their pages (Markdown
editor with live preview rendered by the API), and issue API tokens for MCP
writes.

```sh
cp .env.example .env   # points at the local API on 127.0.0.1:4730
bun run dev            # http://127.0.0.1:4740
bun run build
bun run check
```

Routing is hash based (`#/projects/...`), so the build is one static
`index.html` plus assets. Talks only to the Admin endpoints documented in
`services/api/README.md`.
