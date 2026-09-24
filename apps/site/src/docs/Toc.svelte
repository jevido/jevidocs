<script lang="ts">
  import type { TocItem } from '../lib/api'
  import Icon from '../lib/Icon.svelte'

  let { items, content }: { items: TocItem[]; content?: HTMLElement } = $props()

  let activeIds = $state<string[]>([])

  // Scrollspy: headings currently in the reading band are active; when none
  // are, the last one scrolled past stays active.
  $effect(() => {
    if (!content || items.length === 0) return
    const ids = items.map((i) => decodeURIComponent(i.url.replace(/^#/, '')))
    const els = ids.map((id) => document.getElementById(id)).filter((e): e is HTMLElement => !!e)
    const visible = new Set<string>()
    let last = ids[0]
    const obs = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) visible.add(e.target.id)
          else visible.delete(e.target.id)
        }
        if (visible.size) {
          activeIds = ids.filter((id) => visible.has(id))
          last = activeIds[activeIds.length - 1]
        } else {
          // Pick the last heading above the viewport top.
          const above = els.filter((el) => el.getBoundingClientRect().top < 120)
          activeIds = [above.length ? above[above.length - 1].id : last]
        }
      },
      { rootMargin: '-64px 0px -45% 0px' },
    )
    els.forEach((el) => obs.observe(el))
    return () => obs.disconnect()
  })
</script>

<nav class="toc" aria-label="On this page">
  <p class="heading"><Icon name="text" size={14} /> On this page</p>
  {#if items.length === 0}
    <p class="none">No headings</p>
  {:else}
    <ul>
      {#each items as item (item.url)}
        {@const id = decodeURIComponent(item.url.replace(/^#/, ''))}
        <li style:--indent={Math.max(0, item.depth - 2)}>
          <a href={item.url} class:active={activeIds.includes(id)}>{item.title}</a>
        </li>
      {/each}
    </ul>
  {/if}
</nav>

<style>
  .toc { font-size: 0.85rem; }
  .heading {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0 0 0.75rem;
    color: var(--muted-fg);
    font-size: 0.8rem;
  }
  .none { color: var(--muted-fg); font-size: 0.8rem; margin: 0; }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    border-left: 1px solid var(--border);
  }
  li { padding-left: calc(var(--indent) * 0.85rem); }
  a {
    display: block;
    margin-left: -1px;
    padding: 0.3rem 0 0.3rem 0.85rem;
    border-left: 1px solid transparent;
    color: var(--muted-fg);
    text-decoration: none;
    line-height: 1.35;
    overflow-wrap: anywhere;
    transition: color 0.12s, border-color 0.12s;
  }
  a:hover { color: var(--fg); }
  a.active { color: var(--brand); border-left-color: var(--brand); }
</style>
