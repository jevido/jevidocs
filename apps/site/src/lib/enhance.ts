// Behaviour for the HTML the API renders (see services/api/README.md,
// "Rendered HTML"). The server emits final markup; we only add copy buttons
// and wire up tabs.

import { mountCanvas } from './canvas'

const copyIcon =
  '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>'
const checkIcon =
  '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>'

export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.append(ta)
    ta.select()
    const ok = document.execCommand('copy')
    ta.remove()
    return ok
  }
}

// External links open in a new tab and get a ↗ marker (see prose.css).
function markExternal(root: HTMLElement) {
  for (const a of root.querySelectorAll<HTMLAnchorElement>('a[href^="http"]')) {
    if (a.host === location.host || a.classList.contains('fd-card')) continue
    a.target = '_blank'
    a.rel = 'noopener noreferrer'
    a.classList.add('fd-external')
  }
}

let toastTimer: ReturnType<typeof setTimeout> | undefined

// showToast flashes a short status message at the bottom of the screen.
export function showToast(text: string) {
  let el = document.querySelector<HTMLElement>('.fd-toast')
  if (!el) {
    el = document.createElement('div')
    el.className = 'fd-toast'
    el.setAttribute('role', 'status')
    document.body.append(el)
  }
  el.textContent = text
  el.dataset.show = ''
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => delete el!.dataset.show, 1600)
}

export function enhance(root: HTMLElement) {
  markExternal(root)
  for (const fig of root.querySelectorAll<HTMLElement>('.fd-codeblock')) {
    if (fig.querySelector(':scope > .fd-copy')) continue
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.className = 'fd-copy'
    btn.setAttribute('aria-label', 'Copy code')
    btn.innerHTML = copyIcon
    fig.append(btn)
    if (!fig.querySelector(':scope > .fd-codeblock-title')) fig.dataset.untitled = ''
  }
  // A bare <pre> (no figure) still deserves a copy button.
  for (const pre of root.querySelectorAll<HTMLElement>('pre')) {
    if (pre.closest('.fd-codeblock') || pre.closest('.fd-mermaid')) continue
    const fig = document.createElement('figure')
    fig.className = 'fd-codeblock'
    fig.dataset.untitled = ''
    pre.replaceWith(fig)
    fig.append(pre)
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.className = 'fd-copy'
    btn.setAttribute('aria-label', 'Copy code')
    btn.innerHTML = copyIcon
    fig.append(btn)
  }
  renderMermaid(root)
  for (const tabs of root.querySelectorAll<HTMLElement>('.fd-tabs')) {
    const triggers = tabs.querySelectorAll<HTMLElement>(':scope > .fd-tabs-list > .fd-tab-trigger')
    const panels = tabs.querySelectorAll<HTMLElement>(':scope > .fd-tab')
    const id = `fd-tabs-${++tabsSeq}`
    // ARIA tabs: each trigger controls the panel with the same value.
    triggers.forEach((t, i) => {
      t.setAttribute('role', 'tab')
      t.setAttribute('type', 'button')
      t.id = `${id}-t${i}`
      const panel = [...panels].find((p) => p.dataset.value === t.dataset.tab)
      if (panel) {
        panel.id = `${id}-p${i}`
        panel.setAttribute('role', 'tabpanel')
        panel.setAttribute('aria-labelledby', t.id)
        panel.tabIndex = 0
        t.setAttribute('aria-controls', panel.id)
      }
    })
    const active = [...triggers].find((t) => t.hasAttribute('data-active')) ?? triggers[0]
    if (active) select(tabs, active.dataset.tab ?? '')
  }
}

let tabsSeq = 0

// Arrow keys, Home and End move between tabs (WAI-ARIA tabs pattern).
export function onContentKeydown(e: KeyboardEvent) {
  const trigger = (e.target as Element | null)?.closest<HTMLElement>('.fd-tab-trigger')
  const tabs = trigger?.closest<HTMLElement>('.fd-tabs')
  if (!trigger || !tabs) return
  const all = [...tabs.querySelectorAll<HTMLElement>(':scope > .fd-tabs-list > .fd-tab-trigger')]
  const i = all.indexOf(trigger)
  const next = { ArrowRight: i + 1, ArrowLeft: i - 1, Home: 0, End: all.length - 1 }[e.key]
  if (next === undefined) return
  e.preventDefault()
  const to = all[(next + all.length) % all.length]
  select(tabs, to.dataset.tab ?? '')
  to.focus()
}

function select(tabs: HTMLElement, value: string) {
  for (const t of tabs.querySelectorAll<HTMLElement>(':scope > .fd-tabs-list > .fd-tab-trigger')) {
    const on = t.dataset.tab === value
    t.toggleAttribute('data-active', on)
    t.setAttribute('aria-selected', String(on))
    t.tabIndex = on ? 0 : -1
  }
  for (const p of tabs.querySelectorAll<HTMLElement>(':scope > .fd-tab')) {
    p.toggleAttribute('data-active', p.dataset.value === value)
  }
}

// Delegated click handling: attach once to the content container.
export function onContentClick(e: MouseEvent) {
  const target = e.target as Element | null
  const trigger = target?.closest<HTMLElement>('.fd-tab-trigger')
  if (trigger) {
    const tabs = trigger.closest<HTMLElement>('.fd-tabs')
    if (tabs) select(tabs, trigger.dataset.tab ?? '')
    return
  }
  const anchor = target?.closest<HTMLAnchorElement>('a.fd-anchor')
  if (anchor) {
    // Navigation to the heading still happens; the link is copied as well.
    const url = new URL(anchor.getAttribute('href') ?? '', location.href).href
    copyText(url).then((ok) => ok && showToast('Link copied'))
  }
  const img = target?.closest<HTMLImageElement>('img')
  if (img && !img.closest('a')) {
    zoom(img)
    return
  }
  const copy = target?.closest<HTMLButtonElement>('.fd-copy')
  if (copy) {
    const code = copy.parentElement?.querySelector('pre code, pre')
    const text = code?.textContent ?? ''
    copyText(text).then((ok) => {
      if (!ok) return
      copy.innerHTML = checkIcon
      copy.dataset.copied = ''
      setTimeout(() => {
        copy.innerHTML = copyIcon
        delete copy.dataset.copied
      }, 1500)
    })
  }
}

// ---- Mermaid ---------------------------------------------------------------
// ```mermaid fences arrive as <div class="fd-mermaid"><pre class="fd-mermaid-src">.
// Mermaid (large) is only fetched when a page has a diagram, and diagrams
// are redrawn when the theme flips. On any error the source stays visible.
// Each diagram is drawn at its natural size inside a pan/zoom canvas.
//
// The ELK layout engine (`layout: elk` in a diagram's config) routes edges
// around nodes instead of through them; it is only fetched when a diagram
// asks for it.

type Mermaid = {
  initialize(c: Record<string, unknown>): void
  render(id: string, src: string): Promise<{ svg: string }>
  registerLayoutLoaders(loaders: unknown): void
}

const MERMAID_URL = 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs'
const ELK_URL = 'https://cdn.jsdelivr.net/npm/@mermaid-js/layout-elk@0/dist/mermaid-layout-elk.esm.min.mjs'
let mermaid: Promise<Mermaid> | null = null
let elk: Promise<void> | null = null
let diagramSeq = 0
let observing = false

function loadMermaid(): Promise<Mermaid> {
  mermaid ??= import(/* @vite-ignore */ MERMAID_URL).then((m) => m.default as Mermaid)
  return mermaid
}

function loadElk(m: Mermaid): Promise<void> {
  elk ??= import(/* @vite-ignore */ ELK_URL).then((e) => m.registerLayoutLoaders(e.default))
  return elk
}

// Mermaid's `base` theme with the site's neutral tokens (app.css), so
// diagrams read as part of the page. Mermaid derives shades from these, so
// they must be concrete colours, not CSS variables.
function diagramTheme(dark: boolean) {
  const c = dark
    ? { bg: '#121212', node: '#1a1a1a', border: '#3d3d3d', text: '#ebebeb', muted: '#a3a3a3', line: '#6b6b6b', group: '#161616', shadow: 'rgb(0 0 0 / 0.5)' }
    : { bg: '#fafafa', node: '#ffffff', border: '#d9d9d9', text: '#0a0a0a', muted: '#616161', line: '#a3a3a3', group: '#f3f3f3', shadow: 'rgb(0 0 0 / 0.08)' }
  return {
    theme: 'base',
    fontFamily: 'inherit',
    themeVariables: {
      darkMode: dark,
      background: c.bg,
      fontSize: '14px',
      primaryColor: c.node,
      primaryBorderColor: c.border,
      primaryTextColor: c.text,
      secondaryColor: c.group,
      secondaryBorderColor: c.border,
      secondaryTextColor: c.text,
      tertiaryColor: c.group,
      tertiaryBorderColor: c.border,
      tertiaryTextColor: c.text,
      lineColor: c.line,
      textColor: c.text,
      clusterBkg: c.group,
      clusterBorder: c.border,
      edgeLabelBackground: c.bg,
      noteBkgColor: c.group,
      noteBorderColor: c.border,
      noteTextColor: c.text,
      actorBkg: c.node,
      actorBorder: c.border,
      actorTextColor: c.text,
      actorLineColor: c.line,
      signalColor: c.muted,
      signalTextColor: c.text,
      labelBoxBkgColor: c.node,
      labelBoxBorderColor: c.border,
      labelTextColor: c.text,
      loopTextColor: c.muted,
      activationBkgColor: c.group,
      activationBorderColor: c.border,
      sequenceNumberColor: c.bg,
    },
    // Rounded, lightly lifted nodes; quiet edge labels.
    themeCSS: `
      .node rect.label-container:not([rx]), .node rect.basic:not([rx]),
      rect.actor, .note, .labelBox { rx: 8px; ry: 8px; }
      .cluster rect { rx: 12px; ry: 12px; }
      .node .label-container, rect.actor { filter: drop-shadow(0 1px 2px ${c.shadow}); }
      .node .nodeLabel, .actor { font-weight: 500; }
      .flowchart-link, .relation { stroke-width: 1.5px; }
      .edgeLabel, .edgeLabel p, .labelBkg { background-color: ${c.bg} !important; color: ${c.muted}; font-size: 12px; }
      .edgeLabel rect { fill: ${c.bg}; }
      .cluster-label .nodeLabel { color: ${c.muted}; font-size: 12px; font-weight: 500; }
    `,
  }
}

const isDark = () => document.documentElement.classList.contains('dark')

async function renderMermaid(root: ParentNode) {
  const blocks = [...root.querySelectorAll<HTMLElement>('.fd-mermaid')]
  if (!blocks.length) return
  watchTheme()
  try {
    const m = await loadMermaid()
    const dark = isDark()
    m.initialize({ startOnLoad: false, securityLevel: 'strict', ...diagramTheme(dark) })
    for (const block of blocks) {
      const src = block.querySelector('.fd-mermaid-src')?.textContent ?? ''
      if (!src.trim() || block.dataset.theme === String(dark)) continue
      try {
        if (/\belk\b/.test(src)) await loadElk(m).catch(() => {})
        const { svg } = await m.render(`fd-mermaid-${++diagramSeq}`, src)
        block.querySelector('.fd-mermaid-svg')?.remove()
        const holder = document.createElement('div')
        holder.className = 'fd-mermaid-svg'
        block.prepend(holder)
        mountCanvas(holder, svg)
        block.dataset.rendered = ''
        block.dataset.theme = String(dark)
      } catch {
        delete block.dataset.rendered
      }
    }
  } catch {
    // Offline or blocked CDN: the source stays readable.
  }
}

function watchTheme() {
  if (observing) return
  observing = true
  new MutationObserver(() => renderMermaid(document)).observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
}

// ---- Image zoom ------------------------------------------------------------

function zoom(img: HTMLImageElement) {
  const overlay = document.createElement('div')
  overlay.className = 'fd-zoom'
  overlay.setAttribute('role', 'dialog')
  overlay.setAttribute('aria-label', img.alt || 'Image')
  const big = document.createElement('img')
  big.src = img.currentSrc || img.src
  big.alt = img.alt
  overlay.append(big)
  const close = () => {
    overlay.remove()
    removeEventListener('keydown', onKey)
  }
  const onKey = (e: KeyboardEvent) => {
    if (e.key === 'Escape') close()
  }
  overlay.addEventListener('click', close)
  addEventListener('keydown', onKey)
  document.body.append(overlay)
}
