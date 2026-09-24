import { API_URL } from './api-base'
import { readerAuth } from './reader-auth.svelte'

export { API_URL }

export type Link = { text: string; url: string }

export type Project = {
  slug: string
  name: string
  description: string
  github_url: string
  links: Link[] | null
  public: boolean
  ask?: boolean
  banner?: string
  accent?: string
  logo_url?: string
  version_label?: string
  locales?: string[]
  default_locale?: string
  locale?: string
  versions?: VersionLink[]
  // How the reader got in: public, or a private project through admin
  // rights, membership or a share link.
  access?: 'public' | 'admin' | 'member' | 'domain' | 'share'
  updated_at: string
}

export type VersionLink = { slug: string; name: string; label: string; url: string }

export type PageRef = { type: 'page'; name: string; slug: string; icon?: string }
export type TreeNode =
  | PageRef
  | { type: 'separator'; name: string }
  | {
      type: 'folder'
      name: string
      icon?: string
      index?: PageRef
      children: TreeNode[]
      defaultOpen: boolean
      root?: boolean
      description?: string
    }
export type PageTree = { name: string; children: TreeNode[] }

export type ProjectWithTree = Project & { tree: PageTree }

export type TocItem = { depth: number; title: string; url: string }

export type Page = {
  slug: string
  title: string
  description: string
  icon: string
  html: string
  toc: TocItem[] | null
  breadcrumbs: { name: string; slug?: string }[] | null
  previous: { title: string; slug: string } | null
  next: { title: string; slug: string } | null
  markdown: string
  locale?: string
  fallback?: boolean
  draft?: boolean
  edit_url?: string
  updated_at: string
}

export type SearchResult = {
  type: 'page' | 'heading'
  slug: string
  hash: string
  title: string
  page_title: string
  snippet: string
  fuzzy?: boolean
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
  }
}

async function get<T>(path: string, signal?: AbortSignal, project?: string): Promise<T> {
  const res = await fetch(API_URL + path, {
    signal,
    headers: { Accept: 'application/json', ...readerAuth.headers(project) },
  })
  if (res.status === 401) readerAuth.expire()
  if (!res.ok) {
    let msg = res.statusText
    try {
      const body = await res.json()
      if (body?.error) msg = body.error
    } catch {
      // not JSON
    }
    throw new ApiError(res.status, msg)
  }
  return res.json() as Promise<T>
}

const p = (project: string) => `/api/projects/${encodeURIComponent(project)}`

// The reader's language ('' = project default), added to reads as ?locale=.
let currentLocale = ''
export function setLocale(locale: string) {
  currentLocale = locale
}
const loc = (sep: '?' | '&') => (currentLocale ? `${sep}locale=${encodeURIComponent(currentLocale)}` : '')

export type AskAnswer = { answer: string; sources: { title: string; url: string }[] }

async function post<T>(path: string, body: unknown, project?: string): Promise<T> {
  const res = await fetch(API_URL + path, {
    method: 'POST',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json', ...readerAuth.headers(project) },
    body: JSON.stringify(body),
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new ApiError(res.status, data?.error ?? res.statusText)
  return data as T
}

export const api = {
  ask: (project: string, question: string) => post<AskAnswer>(`${p(project)}/ask`, { question }, project),
  projects: (signal?: AbortSignal) => get<Project[]>('/api/projects', signal),
  project: (project: string, signal?: AbortSignal) => get<ProjectWithTree>(p(project) + loc('?'), signal, project),
  page: (project: string, slug: string, signal?: AbortSignal) =>
    get<Page>(`${p(project)}/page?slug=${encodeURIComponent(slug)}${loc('&')}${previewParam()}`, signal, project),
  search: (project: string, q: string, signal?: AbortSignal) =>
    get<SearchResult[] | null>(`${p(project)}/search?q=${encodeURIComponent(q)}${loc('&')}`, signal, project),
  markdownUrl: (project: string, slug: string) =>
    `${API_URL}${p(project)}/page.md?slug=${encodeURIComponent(slug)}${loc('&')}`,
  // One page view, fire and forget. text/plain avoids a CORS preflight; the
  // API parses the body as JSON regardless.
  trackView: (project: string, slug: string) => {
    try {
      const body = new Blob([JSON.stringify({ slug })], { type: 'text/plain' })
      const url = `${API_URL}${p(project)}/views`
      const headers = readerAuth.headers(project)
      // sendBeacon cannot send headers; private projects need them.
      if (Object.keys(headers).length > 0 || !navigator.sendBeacon?.(url, body)) {
        void fetch(url, { method: 'POST', body, headers, keepalive: true, mode: 'cors' }).catch(() => {})
      }
    } catch {
      // analytics must never break reading
    }
  },
  llmsUrl: (project: string) => `${API_URL}${p(project)}/llms.txt${loc('?')}`,
  llmsFullUrl: (project: string) => `${API_URL}${p(project)}/llms-full.txt${loc('?')}`,
}

// Draft preview links carry ?preview=<token>; pass it on to page reads so an
// unpublished page shows. Read from the current URL at call time.
function previewParam(): string {
  try {
    const t = new URLSearchParams(location.search).get('preview')
    return t ? `&preview=${encodeURIComponent(t)}` : ''
  } catch {
    return ''
  }
}
