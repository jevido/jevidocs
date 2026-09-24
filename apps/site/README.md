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

## Smoke tests

`e2e/smoke.ts` drives Chromium (puppeteer-core) through the API, the docs
reader and, with credentials, the admin. It fails on any failed check or
console error.

```sh
task e2e    # local dev servers (4720, 4740, 4730)
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=admin task e2e
SITE_URL=https://jevidocs.jevido.app ADMIN_URL=https://admin.jevidocs.jevido.app \
  API_URL=https://api.jevidocs.jevido.app task e2e
```

`CHROME_PATH` defaults to `/usr/bin/chromium`.
