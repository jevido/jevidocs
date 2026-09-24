// Hash router: #/projects, #/projects/{slug}, #/projects/{slug}/pages/{id|new}, #/tokens.
export type Route =
  | { name: 'dashboard' }
  | { name: 'projects' }
  | { name: 'project'; slug: string }
  | { name: 'page'; slug: string; id: number | 'new' }
  | { name: 'tokens' }
  | { name: 'notfound' }

function parse(hash: string): Route {
  const parts = hash.replace(/^#\/?/, '').split('/').filter(Boolean).map(decodeURIComponent)
  if (parts.length === 0) return { name: 'dashboard' }
  if (parts[0] === 'tokens' && parts.length === 1) return { name: 'tokens' }
  if (parts[0] === 'projects') {
    if (parts.length === 1) return { name: 'projects' }
    if (parts.length === 2) return { name: 'project', slug: parts[1] }
    if (parts.length === 4 && parts[2] === 'pages') {
      if (parts[3] === 'new') return { name: 'page', slug: parts[1], id: 'new' }
      const id = Number(parts[3])
      if (Number.isInteger(id)) return { name: 'page', slug: parts[1], id }
    }
  }
  return { name: 'notfound' }
}

class Router {
  route = $state<Route>(parse(location.hash))
  /** Return false to cancel navigation (e.g. unsaved changes). */
  guard: (() => boolean) | null = null
  private current = location.hash

  constructor() {
    window.addEventListener('hashchange', () => {
      if (this.guard && !this.guard()) {
        history.replaceState(null, '', this.current || '#/')
        return
      }
      this.guard = null
      this.current = location.hash
      this.route = parse(location.hash)
    })
  }
}

export const router = new Router()

export function go(path: string) {
  location.hash = path
}

export function href(path: string): string {
  return '#' + path
}
