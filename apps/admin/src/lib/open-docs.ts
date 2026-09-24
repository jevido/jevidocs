import { publicDocsUrl, request } from './api.svelte'

// Opens a project's docs in a new tab, signed in as the current admin user:
// the docs site is another origin and cannot see the admin's token, so it
// gets a one-minute handoff code instead. The tab is opened synchronously
// (popup blockers) and pointed at the docs once the code arrives.
export async function openDocs(e: MouseEvent, slug: string) {
  if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return
  e.preventDefault()
  const tab = window.open('about:blank', '_blank')
  let url = publicDocsUrl(slug)
  try {
    const { code } = await request<{ code: string }>('POST', '/api/auth/handoff')
    url += (url.includes('?') ? '&' : '?') + 'handoff=' + encodeURIComponent(code)
  } catch {
    // Not fatal: the reader can still sign in on the docs site.
  }
  if (tab) {
    tab.opener = null
    tab.location.href = url
  } else {
    location.href = url
  }
}
