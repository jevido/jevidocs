// A tiny history-API router for the docs reader. Routes:
//   /docs/<slug...>          -> project "jevidocs"
//   /p/<project>/<slug...>   -> any project
export const DEFAULT_PROJECT = 'jevidocs'

// locale is '' for a project's default language; base then includes the
// locale segment (/p/demo/nl) so every href stays in the reader's language.
export type Route = { project: string; base: string; slug: string; locale: string; root: string }

export function parse(pathname: string): Route {
  const clean = pathname.replace(/\/+$/, '')
  if (clean === '/docs' || clean.startsWith('/docs/')) {
    return { project: DEFAULT_PROJECT, base: '/docs', root: '/docs', locale: '', slug: decode(clean.slice('/docs/'.length)) }
  }
  const m = clean.match(/^\/p\/([^/]+)(?:\/(.*))?$/)
  if (m) return { project: decode(m[1]), base: `/p/${m[1]}`, root: `/p/${m[1]}`, locale: '', slug: decode(m[2] ?? '') }
  return { project: '', base: '/p', root: '/p', locale: '', slug: '' }
}

// withLocale reads a leading locale segment, but only one the project is
// known to have, so a slug like "api" is never mistaken for a language.
export function withLocale(r: Route, locales: string[]): Route {
  const [first, ...rest] = r.slug.split('/')
  if (!first || !locales.includes(first)) return r
  return { ...r, locale: first, base: `${r.root}/${first}`, slug: rest.join('/') }
}

function decode(s: string): string {
  try {
    return decodeURIComponent(s)
  } catch {
    return s
  }
}

export function isDocsPath(pathname: string): boolean {
  return pathname === '/docs' || pathname.startsWith('/docs/') || pathname.startsWith('/p/')
}

class Router {
  pathname = $state(location.pathname)
  hash = $state(location.hash)
  // Non-default locales of the current project, set once it has loaded.
  locales = $state<string[]>([])
  route = $derived(withLocale(parse(this.pathname), this.locales))

  constructor() {
    addEventListener('popstate', () => {
      this.pathname = location.pathname
      this.hash = location.hash
    })
  }

  href(slug: string, base = this.route.base): string {
    return slug ? `${base}/${slug.split('/').map(encodeURIComponent).join('/')}` : base
  }

  navigate(url: string, { replace = false } = {}) {
    const next = new URL(url, location.href)
    const samePage = next.pathname.replace(/\/+$/, '') === this.pathname.replace(/\/+$/, '')
    if (replace) history.replaceState(null, '', next.pathname + next.search + next.hash)
    else history.pushState(null, '', next.pathname + next.search + next.hash)
    this.hash = next.hash
    if (samePage) {
      scrollToHash(next.hash)
    } else {
      this.pathname = next.pathname
    }
  }
}

export function scrollToHash(hash: string) {
  if (!hash) {
    scrollTo({ top: 0 })
    return
  }
  const el = document.getElementById(decodeURIComponent(hash.slice(1)))
  if (el) el.scrollIntoView({ block: 'start' })
}

export const router = new Router()

// Intercepts clicks on same-origin links into the docs reader, including those
// inside rendered Markdown, so navigation never reloads the page.
export function interceptLinks() {
  document.addEventListener('click', (e) => {
    if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return
    const a = (e.target as Element | null)?.closest?.('a')
    if (!a || !a.href || a.target || a.hasAttribute('download')) return
    const url = new URL(a.href)
    if (url.origin !== location.origin || !isDocsPath(url.pathname)) return
    e.preventDefault()
    router.navigate(url.href)
  })
}
