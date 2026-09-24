<script lang="ts">
  import { api, type SearchResult } from '../lib/api'
  import Icon from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'

  let { open = $bindable(false), project }: { open: boolean; project: string } = $props()

  let query = $state('')
  let results = $state<SearchResult[]>([])
  let loading = $state(false)
  let failed = $state(false)
  let selected = $state(0)
  let input = $state<HTMLInputElement>()
  let list = $state<HTMLElement>()

  $effect(() => {
    if (open) {
      queueMicrotask(() => input?.select())
    }
  })

  // Debounced search; stale requests are aborted.
  $effect(() => {
    const q = query.trim()
    if (!open || !q) {
      results = []
      loading = false
      return
    }
    const ctrl = new AbortController()
    loading = true
    const timer = setTimeout(async () => {
      try {
        results = (await api.search(project, q, ctrl.signal)) ?? []
        failed = false
        selected = 0
      } catch (e) {
        if ((e as Error).name !== 'AbortError') {
          failed = true
          results = []
        }
      } finally {
        if (!ctrl.signal.aborted) loading = false
      }
    }, 160)
    return () => {
      clearTimeout(timer)
      ctrl.abort()
    }
  })

  const hrefOf = (r: SearchResult) => router.href(r.slug) + (r.hash ? `#${r.hash}` : '')

  function go(r: SearchResult | undefined) {
    if (!r) return
    open = false
    router.navigate(hrefOf(r))
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      selected = Math.min(selected + 1, results.length - 1)
      scrollSelected()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      selected = Math.max(selected - 1, 0)
      scrollSelected()
    } else if (e.key === 'Enter') {
      e.preventDefault()
      go(results[selected])
    } else if (e.key === 'Escape') {
      open = false
    }
  }

  function scrollSelected() {
    queueMicrotask(() => list?.querySelector('[data-selected]')?.scrollIntoView({ block: 'nearest' }))
  }
</script>

<svelte:window
  onkeydown={(e) => {
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault()
      open = !open
    }
  }} />

{#if open}
  <div class="overlay" role="presentation" onclick={() => (open = false)}></div>
  <div class="dialog" role="dialog" aria-modal="true" aria-label="Search documentation">
    <div class="field">
      <Icon name="search" size={18} />
      <input
        bind:this={input}
        bind:value={query}
        {onkeydown}
        placeholder="Search documentation"
        aria-label="Search"
        autocomplete="off"
        spellcheck="false" />
      <kbd>Esc</kbd>
    </div>
    <div class="results" bind:this={list}>
      {#if !query.trim()}
        <p class="empty">Type to search pages and headings.</p>
      {:else if failed}
        <p class="empty">Search is unavailable right now.</p>
      {:else if !loading && results.length === 0}
        <p class="empty">No results for “{query}”.</p>
      {:else}
        {#each results as r, i (r.slug + '#' + r.hash + i)}
          <a
            href={hrefOf(r)}
            class="result"
            class:heading={r.type === 'heading'}
            data-selected={i === selected ? '' : undefined}
            onmouseenter={() => (selected = i)}
            onclick={(e) => {
              e.preventDefault()
              go(r)
            }}>
            <span class="kind"><Icon name={r.type === 'heading' ? 'hash' : 'file'} size={16} /></span>
            <span class="body">
              <span class="title">{r.title}</span>
              {#if r.type === 'heading'}<span class="page">{r.page_title}</span>{/if}
              {#if r.snippet}<span class="snippet">{@html r.snippet}</span>{/if}
            </span>
          </a>
        {/each}
      {/if}
    </div>
    <div class="footer">
      <span><kbd>↑</kbd> <kbd>↓</kbd> navigate</span>
      <span><kbd>↵</kbd> open</span>
      {#if loading}<span class="spinner" aria-label="Searching"></span>{/if}
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: var(--overlay);
    backdrop-filter: blur(2px);
    animation: fade 0.15s;
  }
  .dialog {
    position: fixed;
    z-index: 61;
    left: 50%;
    top: 12vh;
    width: min(640px, calc(100vw - 2rem));
    transform: translateX(-50%);
    background: var(--popover);
    border: 1px solid var(--border);
    border-radius: 0.9rem;
    box-shadow: var(--shadow);
    overflow: hidden;
    animation: pop 0.16s ease-out;
  }
  .field {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0 1rem;
    border-bottom: 1px solid var(--border);
    color: var(--muted-fg);
  }
  input {
    flex: 1;
    height: 3.25rem;
    border: 0;
    outline: 0;
    background: transparent;
    color: var(--fg);
    font: inherit;
    font-size: 1rem;
  }
  .results {
    max-height: min(420px, 60vh);
    overflow-y: auto;
    padding: 0.4rem;
  }
  .empty {
    margin: 0;
    padding: 2rem 1rem;
    text-align: center;
    color: var(--muted-fg);
    font-size: 0.9rem;
  }
  .result {
    display: flex;
    gap: 0.7rem;
    padding: 0.6rem 0.7rem;
    border-radius: var(--radius);
    text-decoration: none;
    color: var(--fg);
  }
  .result.heading { padding-left: 1.6rem; }
  .result[data-selected] { background: var(--accent); }
  .kind { color: var(--muted-fg); padding-top: 0.15rem; }
  .body { display: flex; flex-direction: column; gap: 0.1rem; min-width: 0; }
  .title { font-size: 0.9rem; font-weight: 500; }
  .page { font-size: 0.75rem; color: var(--muted-fg); }
  .snippet {
    font-size: 0.8rem;
    color: var(--muted-fg);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  .snippet :global(mark) {
    background: transparent;
    color: var(--brand);
    font-weight: 500;
  }
  .footer {
    display: flex;
    gap: 1rem;
    align-items: center;
    padding: 0.55rem 1rem;
    border-top: 1px solid var(--border);
    font-size: 0.75rem;
    color: var(--muted-fg);
    background: var(--card);
  }
  .spinner {
    margin-left: auto;
    width: 14px;
    height: 14px;
    border: 2px solid var(--border);
    border-top-color: var(--fg);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  @keyframes fade { from { opacity: 0; } }
  @keyframes pop { from { opacity: 0; transform: translateX(-50%) scale(0.97); } }
</style>
