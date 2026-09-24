// Asset and user endpoints (see services/api/README.md).
import { API_URL, ApiError, request, session } from './api.svelte'

const enc = encodeURIComponent

export type Asset = { id: number; name: string; url: string; content_type: string; size: number; created_at: string }
export type AdminUser = { id: number; name: string; email: string; role: 'viewer' | 'editor' | 'admin'; created_at: string }

export const ASSET_ACCEPT = 'image/png,image/jpeg,image/gif,image/webp,image/svg+xml,image/avif,application/pdf,text/plain'

async function upload(project: string, file: File): Promise<Asset> {
  const form = new FormData()
  form.append('file', file, file.name)
  const headers: Record<string, string> = { Accept: 'application/json' }
  if (session.token) headers.Authorization = `Bearer ${session.token}`
  let res: Response
  try {
    res = await fetch(`${API_URL}/api/admin/projects/${enc(project)}/assets`, { method: 'POST', headers, body: form })
  } catch {
    throw new ApiError('Cannot reach the API', 0)
  }
  if (res.status === 401) {
    session.set(null)
    throw new ApiError('Session expired, sign in again', 401)
  }
  const data = await res.json().catch(() => null)
  if (!res.ok) throw new ApiError((data as { error?: string } | null)?.error ?? `Upload failed (${res.status})`, res.status)
  return data as Asset
}

/** Markdown that embeds (images) or links (other files) an asset. */
export function assetMarkdown(a: Pick<Asset, 'name' | 'url' | 'content_type'>): string {
  return a.content_type.startsWith('image/') ? `![${a.name}](${a.url})` : `[${a.name}](${a.url})`
}

export const assetsApi = {
  upload,
  list: (project: string) => request<Asset[]>('GET', `/api/admin/projects/${enc(project)}/assets`),
  remove: (project: string, id: number) => request<{ ok: true }>('DELETE', `/api/admin/projects/${enc(project)}/assets/${id}`),
  users: () => request<AdminUser[]>('GET', '/api/admin/users'),
  createUser: (u: { name: string; email: string; password: string; role?: string }) => request<AdminUser>('POST', '/api/admin/users', u),
  setRole: (id: number, role: string) => request<{ ok: true }>('PUT', `/api/admin/users/${id}`, { role }),
  deleteUser: (id: number) => request<{ ok: true }>('DELETE', `/api/admin/users/${id}`),
  changePassword: (current: string, next: string) =>
    request<{ ok: true }>('PUT', '/api/auth/password', { current, new: next }),
}

export function formatSize(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}
