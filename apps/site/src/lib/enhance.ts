// Behaviour for the HTML the API renders (see services/api/README.md,
// "Rendered HTML"). The server emits final markup; we only add copy buttons
// and wire up tabs.

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

export function enhance(root: HTMLElement) {
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
    if (pre.closest('.fd-codeblock')) continue
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

type Mermaid = {
  initialize(c: Record<string, unknown>): void
  render(id: string, src: string): Promise<{ svg: string }>
}

const MERMAID_URL = 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs'
let mermaid: Promise<Mermaid> | null = null
let diagramSeq = 0
let observing = false

function loadMermaid(): Promise<Mermaid> {
  mermaid ??= import(/* @vite-ignore */ MERMAID_URL).then((m) => m.default as Mermaid)
  return mermaid
}

const isDark = () => document.documentElement.classList.contains('dark')

async function renderMermaid(root: ParentNode) {
  const blocks = [...root.querySelectorAll<HTMLElement>('.fd-mermaid')]
  if (!blocks.length) return
  watchTheme()
  try {
    const m = await loadMermaid()
    const dark = isDark()
    m.initialize({ startOnLoad: false, theme: dark ? 'dark' : 'default', securityLevel: 'strict', fontFamily: 'inherit' })
    for (const block of blocks) {
      const src = block.querySelector('.fd-mermaid-src')?.textContent ?? ''
      if (!src.trim() || block.dataset.theme === String(dark)) continue
      try {
        const { svg } = await m.render(`fd-mermaid-${++diagramSeq}`, src)
        block.querySelector('.fd-mermaid-svg')?.remove()
        const holder = document.createElement('div')
        holder.className = 'fd-mermaid-svg'
        holder.innerHTML = svg
        block.prepend(holder)
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
