import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig, type Connect, type Plugin } from 'vite'
import { resolve } from 'node:path'

// The docs reader is a single-page app under /docs/* and /p/*. In production
// nginx falls back to their index.html; this does the same for dev/preview.
function docsFallback(): Plugin {
  const rewrite: Connect.NextHandleFunction = (req, _res, next) => {
    const url = req.url ?? '/'
    const path = url.split('?')[0]
    const isFile = /\.[a-z0-9]+$/i.test(path)
    if (!isFile) {
      if (path === '/docs' || path.startsWith('/docs/')) req.url = '/docs/index.html'
      else if (path === '/p' || path.startsWith('/p/')) req.url = '/p/index.html'
    }
    next()
  }
  return {
    name: 'docs-spa-fallback',
    configureServer(server) {
      server.middlewares.use(rewrite)
    },
    configurePreviewServer(server) {
      server.middlewares.use(rewrite)
    },
  }
}

// Ports follow the repo's 47xx scheme (see CLAUDE.md).
export default defineConfig({
  plugins: [docsFallback(), svelte()],
  build: {
    rollupOptions: {
      input: {
        home: resolve(import.meta.dirname, 'index.html'),
        docs: resolve(import.meta.dirname, 'docs/index.html'),
        project: resolve(import.meta.dirname, 'p/index.html'),
      },
    },
  },
  server: { host: '127.0.0.1', port: 4720, strictPort: true },
  preview: { host: '127.0.0.1', port: 4721, strictPort: true },
})
