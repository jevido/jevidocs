export type User = { id: number; name: string; email: string; role?: 'viewer' | 'editor' | 'admin' }

export type Link = { text: string; url: string }

export type Project = {
  slug: string
  name: string
  description: string
  github_url: string
  links: Link[]
  public: boolean
  edit_url: string
  banner: string
  accent: string
  logo_url: string
  version_group?: string
  version_label?: string
  locales?: string[]
  default_locale?: string
  managed?: boolean
  updated_at: string
}

export type ProjectInput = Omit<Project, 'updated_at' | 'managed' | 'locales'> & {
  // Comma list ("en,nl"); empty for a single-language project.
  locales?: string
}

export type AdminPageInput = {
  slug: string
  title: string
  description: string
  icon: string
  position: number
  section: string
  published: boolean
  body: string
  // '' = the project's default language.
  locale?: string
}

export type AdminPage = AdminPageInput & {
  id: number
  // 'openapi': the body is an OpenAPI document, shown as an API reference.
  kind?: '' | 'openapi'
  project: string
  updated_by?: string
  updated_at: string
  created_at: string
}

export type TocItem = { depth: number; title: string; url: string }

export type Token = { id: number; name: string; last_used_at: string | null; created_at: string }

export type Stats = { projects: number; pages: number; tokens: number; views_30d?: number }
