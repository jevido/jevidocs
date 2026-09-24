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
  for (const tabs of root.querySelectorAll<HTMLElement>('.fd-tabs')) {
    const triggers = tabs.querySelectorAll<HTMLElement>(':scope > .fd-tabs-list > .fd-tab-trigger')
    if (triggers.length && ![...triggers].some((t) => t.hasAttribute('data-active'))) {
      select(tabs, triggers[0].dataset.tab ?? '')
    }
    triggers.forEach((t) => {
      t.setAttribute('role', 'tab')
      t.setAttribute('type', 'button')
      t.setAttribute('aria-selected', String(t.hasAttribute('data-active')))
    })
  }
}

function select(tabs: HTMLElement, value: string) {
  for (const t of tabs.querySelectorAll<HTMLElement>(':scope > .fd-tabs-list > .fd-tab-trigger')) {
    const on = t.dataset.tab === value
    t.toggleAttribute('data-active', on)
    t.setAttribute('aria-selected', String(on))
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
