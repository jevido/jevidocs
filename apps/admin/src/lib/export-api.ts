import { API_URL, request, session } from './api.svelte'

/** Downloads the project as a zip of Markdown files. */
export async function downloadExport(project: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/admin/projects/${encodeURIComponent(project)}/export`, {
    headers: session.token ? { Authorization: `Bearer ${session.token}` } : {},
  })
  if (!res.ok) throw new Error(`Export failed (${res.status})`)
  const url = URL.createObjectURL(await res.blob())
  const a = document.createElement('a')
  a.href = url
  a.download = `${project}.zip`
  a.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}

/** A 7-day link that shows the page on the docs site even unpublished. */
export function previewLink(project: string, id: number): Promise<{ url: string }> {
  return request('POST', `/api/admin/projects/${encodeURIComponent(project)}/pages/${id}/preview-link`)
}
