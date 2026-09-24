<script lang="ts">
  import Icon from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import { sidebar } from '../lib/sidebar.svelte'

  // Keyboard shortcuts ([ / ] and Alt+←/→ for previous/next), the back-to-top
  // button and the button that reopens a collapsed sidebar.
  let { previous, next }: { previous?: string | null; next?: string | null } = $props()

  let scrolled = $state(false)

  function typing(t: EventTarget | null): boolean {
    const el = t as HTMLElement | null
    return !!el && (el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName))
  }

  function onkeydown(e: KeyboardEvent) {
    if (typing(e.target) || e.metaKey || e.ctrlKey) return
    if (document.querySelector('[role="dialog"]')) return
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
  @media print {
    .float { display: none; }
  }
</style>
