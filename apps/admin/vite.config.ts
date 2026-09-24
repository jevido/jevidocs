import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'

// Ports follow the repo's 47xx scheme (see CLAUDE.md).
export default defineConfig({
  plugins: [svelte()],
  server: { host: '127.0.0.1', port: 4740, strictPort: true },
  preview: { host: '127.0.0.1', port: 4741, strictPort: true },
})
