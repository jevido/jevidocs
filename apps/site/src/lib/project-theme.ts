// Applies a project's accent (a preset name or a CSS colour) and page meta
// tags. The API validates accents strictly before they get here.

const PRESETS = new Set(['neutral', 'ocean', 'purple', 'emerald', 'ruby'])

export function applyAccent(accent: string | undefined | null): void {
  const root = document.documentElement
  root.removeAttribute('data-preset')
  root.style.removeProperty('--brand')
  const a = (accent ?? '').trim()
  if (!a) return
  if (PRESETS.has(a.toLowerCase())) root.setAttribute('data-preset', a.toLowerCase())
  else if (CSS.supports('color', a)) root.style.setProperty('--brand', a)
}

function meta(attr: 'name' | 'property', key: string, value: string): void {
  let el = document.head.querySelector<HTMLMetaElement>(`meta[${attr}="${key}"]`)
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  el.content = value
}

export function setPageMeta(title: string, description: string, url: string): void {
  meta('name', 'description', description)
  meta('property', 'og:title', title)
  meta('property', 'og:description', description)
  meta('property', 'og:type', 'article')
  meta('property', 'og:url', url)
  let link = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'canonical'
    document.head.appendChild(link)
  }
  link.href = url
}
