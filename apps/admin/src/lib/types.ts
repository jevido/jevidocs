export type User = { id: number; name: string; email: string }

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
  managed?: boolean
  updated_at: string
}

export type ProjectInput = Omit<Project, 'updated_at' | 'managed'>

export type AdminPageInput = {
  slug: string
  title: string
  description: string
  icon: string
  position: number
  section: string
  published: boolean
  body: string
}

export type AdminPage = AdminPageInput & {
  id: number
  project: string
  updated_at: string
  created_at: string
}

export type TocItem = { depth: number; title: string; url: string }

export type Token = { id: number; name: string; last_used_at: string | null; created_at: string }

export type Stats = { projects: number; pages: number; tokens: number }
