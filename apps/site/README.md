# site

jevidocs.jevido.app: the landing page and the docs reader. Svelte 5 + Vite +
TypeScript, no SvelteKit.

| Entry | Path | What |
| ----- | ---- | ---- |
| `index.html` → `src/home.ts` | `/` | Landing page |
| `docs/index.html` → `src/docs.ts` | `/docs/*` | Docs reader for the `jevidocs` project |
| `p/index.html` → `src/docs.ts` | `/p/{project}/*` | Docs reader for any project |

The reader is a single-page app with a small history-API router
(`src/lib/router.svelte.ts`). Hosting must fall back to `docs/index.html` for
`/docs/*` and to `p/index.html` for `/p/*` (the Vite dev/preview servers do
this themselves). Content, page trees and search come from the API
(`services/api/README.md`); the API renders Markdown to HTML and the site only
styles it (`src/prose.css`) and adds copy buttons and tabs (`src/lib/enhance.ts`).

```sh
bun run dev      # http://127.0.0.1:4720, talks to VITE_API_URL
bun run build    # dist/
bun run check
```

`VITE_API_URL` defaults to `https://api.jevidocs.jevido.app`; for a local API
use `VITE_API_URL=http://127.0.0.1:4730 bun run dev`.
