export const API_URL: string = (import.meta.env.VITE_API_URL ?? 'https://api.jevidocs.jevido.app').replace(/\/$/, '')

export type Link = { text: string; url: string }

export type Project = {
  slug: string
  name: string
  description: string
  github_url: string
  links: Link[] | null
  public: boolean
  banner?: string
  updated_at: string
}

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
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
  }
}

async function get<T>(path: string, signal?: AbortSignal): Promise<T> {
  const res = await fetch(API_URL + path, { signal, headers: { Accept: 'application/json' } })
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

export const api = {
  projects: (signal?: AbortSignal) => get<Project[]>('/api/projects', signal),
  project: (project: string, signal?: AbortSignal) => get<ProjectWithTree>(p(project), signal),
  page: (project: string, slug: string, signal?: AbortSignal) =>
    get<Page>(`${p(project)}/page?slug=${encodeURIComponent(slug)}`, signal),
  search: (project: string, q: string, signal?: AbortSignal) =>
    get<SearchResult[] | null>(`${p(project)}/search?q=${encodeURIComponent(q)}`, signal),
  markdownUrl: (project: string, slug: string) => `${API_URL}${p(project)}/page.md?slug=${encodeURIComponent(slug)}`,
  llmsUrl: (project: string) => `${API_URL}${p(project)}/llms.txt`,
  llmsFullUrl: (project: string) => `${API_URL}${p(project)}/llms-full.txt`,
}
