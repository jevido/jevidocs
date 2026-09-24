<script lang="ts">
  import Icon from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import { sidebar } from '../lib/sidebar.svelte'

  // Keyboard shortcuts ([ / ] and Alt+←/→ for previous/next), the back-to-top
  // button and the button that reopens a collapsed sidebar.
  let {
    previous,
    next,
    onsearch,
  }: { previous?: string | null; next?: string | null; onsearch?: () => void } = $props()

  let scrolled = $state(false)
  let help = $state(false)

  const shortcuts: [string, string][] = [
    ['⌘K / Ctrl+K', 'Search'],
    ['/', 'Search'],
    ['[  or  Alt+←', 'Previous page'],
    ['] or  Alt+→', 'Next page'],
    ['?', 'Show this help'],
    ['Esc', 'Close dialogs'],
  ]

  function typing(t: EventTarget | null): boolean {
    const el = t as HTMLElement | null
    return !!el && (el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName))
  }

  function onkeydown(e: KeyboardEvent) {
    if (typing(e.target) || e.metaKey || e.ctrlKey) return
    if (help && e.key === 'Escape') {
      help = false
      return
    }
    if (document.querySelector('[role="dialog"]')) return
    if (e.key === '/' && onsearch) {
      e.preventDefault()
      onsearch()
      return
    }
    if (e.key === '?') {
      e.preventDefault()
      help = true
      return
    }
    const back = (e.key === '[' && !e.altKey) || (e.altKey && e.key === 'ArrowLeft')
    const fwd = (e.key === ']' && !e.altKey) || (e.altKey && e.key === 'ArrowRight')
    const slug = back ? previous : fwd ? next : undefined
    if (slug === undefined || slug === null) return
    e.preventDefault()
    router.navigate(router.href(slug))
  }
</script>

<svelte:window {onkeydown} onscroll={() => (scrolled = scrollY > 600)} />

{#if sidebar.collapsed}
  <button class="float reopen" type="button" aria-label="Show sidebar" title="Show sidebar" onclick={() => sidebar.toggle(false)}>
    <Icon name="menu" size={16} />
  </button>
{/if}

{#if help}
  <div class="help-scrim" role="presentation" onclick={() => (help = false)}></div>
  <div class="help" role="dialog" aria-modal="true" aria-labelledby="kbd-help-title">
    <h2 id="kbd-help-title">Keyboard shortcuts</h2>
    <dl>
      {#each shortcuts as [keys, what] (keys)}
        <dt><kbd>{keys}</kbd></dt>
        <dd>{what}</dd>
      {/each}
    </dl>
    <!-- svelte-ignore a11y_autofocus -->
    <button type="button" class="close" autofocus onclick={() => (help = false)}>Close</button>
  </div>
{/if}

{#if scrolled}
  <button class="float top" type="button" aria-label="Back to top" title="Back to top" onclick={() => scrollTo({ top: 0, behavior: 'smooth' })}>
    <Icon name="chevronDown" size={16} />
  </button>
{/if}

<style>
  .float {
    position: fixed;
    z-index: 45;
    display: grid;
    place-items: center;
    width: 2.25rem;
    height: 2.25rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--popover);
    color: var(--fg);
    box-shadow: var(--shadow);
    cursor: pointer;
  }
  .float:hover { background: var(--accent); }
  .reopen { left: 1rem; bottom: 1rem; }
  .top { right: 1.25rem; bottom: 1.25rem; }
  .top :global(svg) { transform: rotate(180deg); }
  @media (max-width: 800px) {
    .reopen { display: none; }
  }
  .help-scrim { position: fixed; inset: 0; z-index: 60; background: var(--overlay); }
  .help {
    position: fixed;
    z-index: 61;
    top: 20vh;
    left: 50%;
    transform: translateX(-50%);
    width: min(24rem, calc(100vw - 2rem));
    padding: 1.1rem 1.25rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--popover);
    box-shadow: var(--shadow);
  }
  .help h2 { margin: 0 0 0.75rem; font-size: 1rem; }
  .help dl { display: grid; grid-template-columns: auto 1fr; gap: 0.45rem 1rem; margin: 0 0 1rem; font-size: 0.875rem; }
  .help dt, .help dd { margin: 0; }
  .help dd { color: var(--muted-fg); }
  .help kbd {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    padding: 0.1rem 0.35rem;
    border: 1px solid var(--border);
    border-radius: 0.3rem;
    background: var(--muted);
    white-space: pre;
  }
  .close {
    font: inherit;
    font-size: 0.85rem;
    padding: 0.35rem 0.8rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
    color: var(--fg);
    cursor: pointer;
  }
  @media print {
    .float { display: none; }
  }
</style>
