import { mount } from 'svelte'
import './app.css'
import './prose.css'
import DocsApp from './docs/DocsApp.svelte'
import ProjectDirectory from './docs/ProjectDirectory.svelte'
import { readerAuth } from './lib/reader-auth.svelte'
import { parse } from './lib/router.svelte'

// Sign in first when arriving from the admin (?handoff=), so a private
// project loads on the first request instead of flashing "not found".
void readerAuth.captureHandoff().finally(() => {
  // /p and /p/ list the public projects; everything else is the reader.
  const target = document.getElementById('app')!
  if (parse(location.pathname).project === '') {
    mount(ProjectDirectory, { target })
  } else {
    mount(DocsApp, { target })
  }
})
