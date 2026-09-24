import type {
  AdminPage,
  AdminPageInput,
  Project,
  ProjectInput,
  Stats,
  TocItem,
  Token,
  User,
} from './types'

export const API_URL: string = (import.meta.env.VITE_API_URL ?? 'https://api.jevidocs.jevido.app').replace(/\/$/, '')
export const SITE_URL = 'https://jevidocs.jevido.app'

const TOKEN_KEY = 'jevidocs.admin.token'

export function publicDocsUrl(slug: string): string {
  return slug === 'jevidocs' ? `${SITE_URL}/docs` : `${SITE_URL}/p/${slug}`
}

function readToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

class Session {
  token = $state<string | null>(readToken())
  user = $state<User | null>(null)

  set(token: string | null, user: User | null = null) {
    this.token = token
    this.user = user
    try {
      if (token) localStorage.setItem(TOKEN_KEY, token)
      else localStorage.removeItem(TOKEN_KEY)
    } catch {
      // Storage blocked: the session lasts until reload.
    }
  }
}

export const session = new Session()

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
  }
}

export async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (session.token) headers.Authorization = `Bearer ${session.token}`

  let res: Response
  try {
    res = await fetch(API_URL + path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError('Cannot reach the API', 0)
  }

  if (res.status === 401 && path !== '/api/auth/login') {
    session.set(null)
    throw new ApiError('Session expired, sign in again', 401)
  }

  const text = await res.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = text
    }
  }
  if (!res.ok) {
    const msg =
      data && typeof data === 'object' && 'error' in data
        ? String((data as { error: unknown }).error)
        : `Request failed (${res.status})`
    throw new ApiError(msg, res.status)
  }
  return data as T
}

const enc = encodeURIComponent

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string; user: User }>('POST', '/api/auth/login', { email, password }),
  me: () => request<{ user: User }>('GET', '/api/auth/me'),
  logout: () => request<{ ok: boolean }>('POST', '/api/auth/logout'),

  stats: () => request<Stats>('GET', '/api/admin/stats'),

  projects: () => request<Project[]>('GET', '/api/admin/projects'),
  createProject: (p: ProjectInput) => request<Project>('POST', '/api/admin/projects', p),
  updateProject: (slug: string, p: ProjectInput) =>
    request<Project>('PUT', `/api/admin/projects/${enc(slug)}`, p),
  deleteProject: (slug: string) => request<{ ok: boolean }>('DELETE', `/api/admin/projects/${enc(slug)}`),

  pages: (project: string) => request<AdminPage[]>('GET', `/api/admin/projects/${enc(project)}/pages`),
  page: (project: string, id: number) =>
    request<AdminPage>('GET', `/api/admin/projects/${enc(project)}/pages/${id}`),
  createPage: (project: string, p: AdminPageInput) =>
    request<AdminPage>('POST', `/api/admin/projects/${enc(project)}/pages`, p),
  updatePage: (project: string, id: number, p: AdminPageInput) =>
    request<AdminPage>('PUT', `/api/admin/projects/${enc(project)}/pages/${id}`, p),
  deletePage: (project: string, id: number) =>
    request<{ ok: boolean }>('DELETE', `/api/admin/projects/${enc(project)}/pages/${id}`),

  importOpenAPI: (project: string, spec: string, prefix: string) =>
    request<{ created: number; updated: number; unchanged: number; deleted: number }>(
      'POST',
      `/api/admin/projects/${enc(project)}/openapi`,
      { spec, prefix },
    ),
  preview: (body: string) => request<{ html: string; toc: TocItem[] }>('POST', '/api/admin/preview', { body }),

  tokens: () => request<Token[]>('GET', '/api/admin/tokens'),
  createToken: (name: string) => request<{ id: number; name: string; token: string }>('POST', '/api/admin/tokens', { name }),
  deleteToken: (id: number) => request<{ ok: boolean }>('DELETE', `/api/admin/tokens/${id}`),
}
