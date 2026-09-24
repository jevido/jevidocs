import { request } from './api.svelte'

export type Member = { user_id: number; name: string; email: string; role: 'viewer' | 'editor' | 'admin' }
export type ShareLink = {
  id: number
  name: string
  expires_at: string | null
  last_used_at: string | null
  created_at: string
}
export type NewShareLink = { id: number; name: string; token: string; url: string }

const base = (project: string) => `/api/admin/projects/${encodeURIComponent(project)}`

export const accessApi = {
  members: (project: string) => request<Member[]>('GET', `${base(project)}/members`),
  addMember: (project: string, userId: number) =>
    request<{ ok: boolean }>('POST', `${base(project)}/members`, { user_id: userId }),
  removeMember: (project: string, userId: number) =>
    request<{ ok: boolean }>('DELETE', `${base(project)}/members/${userId}`),
  domains: (project: string) => request<{ domains: string[] }>('GET', `${base(project)}/domains`),
  setDomains: (project: string, domains: string[]) =>
    request<{ domains: string[] }>('PUT', `${base(project)}/domains`, { domains }),
  shares: (project: string) => request<ShareLink[]>('GET', `${base(project)}/shares`),
  createShare: (project: string, name: string, expiresDays: number) =>
    request<NewShareLink>('POST', `${base(project)}/shares`, { name, expires_days: expiresDays }),
  revokeShare: (project: string, id: number) => request<{ ok: boolean }>('DELETE', `${base(project)}/shares/${id}`),
}
