// Pan/zoom canvas for rendered diagrams. The SVG keeps its natural size and
// the viewport moves over it, so a large diagram stays legible instead of
// being shrunk to the column width.
//
// Drag pans, Ctrl/⌘ + wheel (or a trackpad pinch, which browsers report as
// ctrl + wheel) zooms, two fingers pinch on touch screens. A plain wheel is
// left to the page so a diagram never traps scrolling; in fullscreen the
// wheel pans instead.

type View = {
  stage: HTMLElement
  hint: HTMLElement
  w: number // natural diagram size
  h: number
  x: number
  y: number
  k: number
  touched: boolean // the reader moved it; stop refitting on resize
}

const MIN_ZOOM = 0.1
const MAX_ZOOM = 8
const PAD = 16
const GRID = 20
const views = new WeakMap<HTMLElement, View>()

const icon = (d: string) =>
  `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${d}</svg>`
const icons = {
  in: icon('<path d="M5 12h14"/><path d="M12 5v14"/>'),
  out: icon('<path d="M5 12h14"/>'),
  fit: icon('<path d="M3 7V5a2 2 0 0 1 2-2h2"/><path d="M17 3h2a2 2 0 0 1 2 2v2"/><path d="M21 17v2a2 2 0 0 1-2 2h-2"/><path d="M7 21H5a2 2 0 0 1-2-2v-2"/>'),
  full: icon('<path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/>'),
}

let resize: ResizeObserver | null = null

function observe(vp: HTMLElement) {
  resize ??= new ResizeObserver((entries) => {
    for (const e of entries) {
      const el = e.target as HTMLElement
      // Theme changes and navigation replace the canvas; drop the old one.
      if (!el.isConnected) {
        resize!.unobserve(el)
        continue
      }
      const v = views.get(el)
      if (v && !v.touched) fit(el)
    }
  })
  resize.observe(vp)
}

// mountCanvas replaces the contents of holder with a pan/zoom viewport
// around the SVG markup.
export function mountCanvas(holder: HTMLElement, svgMarkup: string) {
  const vp = document.createElement('div')
  vp.className = 'fd-canvas'
  vp.tabIndex = 0
  vp.setAttribute('role', 'group')
  vp.setAttribute('aria-roledescription', 'canvas')
  vp.setAttribute('aria-label', 'Diagram. Drag to pan, Ctrl + scroll or + and − to zoom, 0 to fit.')

  const stage = document.createElement('div')
  stage.className = 'fd-canvas-stage'
  stage.innerHTML = svgMarkup
  const svg = stage.querySelector('svg')
  if (!svg) {
    holder.replaceChildren(stage)
    return
  }
  // Mermaid emits width="100%" with a max-width; pin the natural size.
  const vb = svg.viewBox.baseVal
  const w = vb?.width || svg.getBoundingClientRect().width || 400
  const h = vb?.height || svg.getBoundingClientRect().height || 300
  svg.setAttribute('width', String(w))
  svg.setAttribute('height', String(h))
  svg.style.maxWidth = 'none'

  const hint = document.createElement('div')
  hint.className = 'fd-canvas-hint'
  hint.setAttribute('aria-hidden', 'true')
  hint.textContent = `${/Mac|iP(hone|ad)/.test(navigator.platform) ? '⌘' : 'Ctrl'} + scroll to zoom`

  const controls = document.createElement('div')
  controls.className = 'fd-canvas-controls'
  const button = (label: string, html: string, act: string) => {
    const b = document.createElement('button')
    b.type = 'button'
    b.className = 'fd-canvas-btn'
    b.dataset.act = act
    b.setAttribute('aria-label', label)
    b.title = label
    b.innerHTML = html
    controls.append(b)
  }
  button('Zoom in', icons.in, 'in')
  button('Zoom out', icons.out, 'out')
  button('Fit to view', icons.fit, 'fit')
  if (document.fullscreenEnabled) button('Fullscreen', icons.full, 'full')

  vp.append(stage, hint, controls)
  holder.replaceChildren(vp)
  views.set(vp, { stage, hint, w, h, x: 0, y: 0, k: 1, touched: false })
  wire(vp)
  fit(vp)
  observe(vp)
}

// ---- view ------------------------------------------------------------------

function apply(vp: HTMLElement) {
  const v = views.get(vp)!
  v.stage.style.transform = `translate(${v.x}px, ${v.y}px) scale(${v.k})`
  // The dot grid moves with the diagram so panning reads as panning.
  vp.style.backgroundSize = `${GRID * v.k}px ${GRID * v.k}px`
  vp.style.backgroundPosition = `${v.x}px ${v.y}px`
}

// fit sizes the viewport to the diagram (within limits) and centres the
// whole diagram in it, never enlarging past 100%.
function fit(vp: HTMLElement) {
  const v = views.get(vp)!
  const full = document.fullscreenElement?.contains(vp) ?? false
  const W = vp.clientWidth
  if (!full) {
    const k = Math.min(1, (W - 2 * PAD) / v.w)
    const max = Math.max(240, innerHeight * 0.7)
    vp.style.setProperty('--fd-canvas-h', `${Math.round(Math.min(max, Math.max(192, v.h * k + 2 * PAD)))}px`)
  }
  const H = vp.clientHeight
  v.k = Math.max(MIN_ZOOM, Math.min(1, (W - 2 * PAD) / v.w, (H - 2 * PAD) / v.h))
  v.x = (W - v.w * v.k) / 2
  v.y = (H - v.h * v.k) / 2
  v.touched = false
  apply(vp)
}

function zoomAt(vp: HTMLElement, factor: number, cx: number, cy: number) {
  const v = views.get(vp)!
  const k = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, v.k * factor))
  v.x = cx - ((cx - v.x) * k) / v.k
  v.y = cy - ((cy - v.y) * k) / v.k
  v.k = k
  v.touched = true
  apply(vp)
}

function pan(vp: HTMLElement, dx: number, dy: number) {
  const v = views.get(vp)!
  v.x += dx
  v.y += dy
  v.touched = true
  apply(vp)
}

const centre = (vp: HTMLElement) => [vp.clientWidth / 2, vp.clientHeight / 2] as const

// ---- input -----------------------------------------------------------------

function wire(vp: HTMLElement) {
  const v = views.get(vp)!
  const local = (e: { clientX: number; clientY: number }) => {
    const r = vp.getBoundingClientRect()
    return [e.clientX - r.left, e.clientY - r.top] as const
  }

  vp.addEventListener('click', (e) => {
    const b = (e.target as Element).closest<HTMLElement>('.fd-canvas-btn')
    if (!b) return
    const [cx, cy] = centre(vp)
    if (b.dataset.act === 'in') zoomAt(vp, 1.25, cx, cy)
    else if (b.dataset.act === 'out') zoomAt(vp, 0.8, cx, cy)
    else if (b.dataset.act === 'fit') fit(vp)
    else if (b.dataset.act === 'full') {
      const target = vp.closest<HTMLElement>('.fd-mermaid') ?? vp
      if (document.fullscreenElement) document.exitFullscreen()
      else target.requestFullscreen().catch(() => {})
    }
  })

  vp.addEventListener('dblclick', (e) => {
    if ((e.target as Element).closest('.fd-canvas-controls')) return
    const [cx, cy] = local(e)
    zoomAt(vp, e.shiftKey ? 0.5 : 2, cx, cy)
  })

  let hintTimer: ReturnType<typeof setTimeout> | undefined
  vp.addEventListener(
    'wheel',
    (e) => {
      const full = document.fullscreenElement?.contains(vp) ?? false
      const scale = e.deltaMode === 1 ? 16 : e.deltaMode === 2 ? vp.clientHeight : 1
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault()
        const d = Math.max(-30, Math.min(30, e.deltaY * scale))
        const [cx, cy] = local(e)
        zoomAt(vp, Math.exp(-d * 0.01), cx, cy)
      } else if (full) {
        e.preventDefault()
        pan(vp, -e.deltaX * scale, -e.deltaY * scale)
      } else {
        v.hint.dataset.show = ''
        clearTimeout(hintTimer)
        hintTimer = setTimeout(() => delete v.hint.dataset.show, 1200)
      }
    },
    { passive: false },
  )

  // One pointer pans; two pointers pinch around their midpoint.
  const pointers = new Map<number, readonly [number, number]>()
  let pinch: { dist: number; mid: readonly [number, number] } | null = null
  const pinchState = () => {
    const [a, b] = [...pointers.values()]
    return { dist: Math.hypot(a[0] - b[0], a[1] - b[1]), mid: [(a[0] + b[0]) / 2, (a[1] + b[1]) / 2] as const }
  }

  vp.addEventListener('pointerdown', (e) => {
    if (e.button !== 0 || (e.target as Element).closest('.fd-canvas-controls')) return
    vp.setPointerCapture(e.pointerId)
    pointers.set(e.pointerId, local(e))
    if (pointers.size === 2) pinch = pinchState()
    vp.dataset.dragging = ''
  })
  vp.addEventListener('pointermove', (e) => {
    const prev = pointers.get(e.pointerId)
    if (!prev) return
    const now = local(e)
    pointers.set(e.pointerId, now)
    if (pointers.size === 1) {
      pan(vp, now[0] - prev[0], now[1] - prev[1])
    } else if (pointers.size === 2 && pinch) {
      const next = pinchState()
      pan(vp, next.mid[0] - pinch.mid[0], next.mid[1] - pinch.mid[1])
      if (pinch.dist > 0) zoomAt(vp, next.dist / pinch.dist, next.mid[0], next.mid[1])
      pinch = next
    }
  })
  const release = (e: PointerEvent) => {
    pointers.delete(e.pointerId)
    pinch = pointers.size === 2 ? pinchState() : null
    if (!pointers.size) delete vp.dataset.dragging
  }
  vp.addEventListener('pointerup', release)
  vp.addEventListener('pointercancel', release)

  vp.addEventListener('keydown', (e) => {
    if (e.target !== vp) return
    const [cx, cy] = centre(vp)
    const step = 48
    const acts: Record<string, () => void> = {
      '+': () => zoomAt(vp, 1.25, cx, cy),
      '=': () => zoomAt(vp, 1.25, cx, cy),
      '-': () => zoomAt(vp, 0.8, cx, cy),
      '0': () => fit(vp),
      ArrowLeft: () => pan(vp, step, 0),
      ArrowRight: () => pan(vp, -step, 0),
      ArrowUp: () => pan(vp, 0, step),
      ArrowDown: () => pan(vp, 0, -step),
    }
    const act = acts[e.key]
    if (!act || e.ctrlKey || e.metaKey || e.altKey) return
    e.preventDefault()
    act()
  })
}

// Entering or leaving fullscreen changes the viewport size completely.
document.addEventListener('fullscreenchange', () => {
  for (const vp of document.querySelectorAll<HTMLElement>('.fd-canvas')) {
    if (views.has(vp)) requestAnimationFrame(() => fit(vp))
  }
})
