<script lang="ts">
  import { api, type AskAnswer, type SearchResult } from '../lib/api'
  import { miniMarkdown } from '../lib/mini-md'
  import Icon from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'

  let {
    open = $bindable(false),
    project,
    ask = false,
  }: { open: boolean; project: string; ask?: boolean } = $props()

  // Ask AI (only when the API has a model configured).
  let answer = $state<AskAnswer | null>(null)
  let asking = $state(false)
  let askError = $state('')
  async function askAI() {
    const q = query.trim()
    if (!q || asking) return
    asking = true
    askError = ''
    answer = null
    try {
      answer = await api.ask(project, q)
    } catch (e) {
      askError = e instanceof Error ? e.message : String(e)
    } finally {
      asking = false
    }
  }
  $effect(() => {
    void query
    answer = null
    askError = ''
  })

  let query = $state('')
  let results = $state<SearchResult[]>([])
  let loading = $state(false)
  let failed = $state(false)
  let selected = $state(0)
  let input = $state<HTMLInputElement>()
  let list = $state<HTMLElement>()

  // Keep Tab inside the dialog while it is open.
  function trapFocus(e: KeyboardEvent) {
    if (e.key !== 'Tab') return
    const box = e.currentTarget as HTMLElement
    const items = [...box.querySelectorAll<HTMLElement>('input, button, a[href], [tabindex]:not([tabindex="-1"])')]
    if (!items.length) return
    const first = items[0]
    const last = items[items.length - 1]
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault()
      last.focus()
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault()
      first.focus()
    }
  }

  // Focus the input on open and give focus back to the opener on close.
  let returnTo: HTMLElement | null = null
  $effect(() => {
    if (open) {
      returnTo = document.activeElement as HTMLElement | null
      queueMicrotask(() => input?.select())
    } else if (returnTo) {
      returnTo.focus?.()
      returnTo = null
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
  <div class="dialog" role="dialog" aria-modal="true" aria-label="Search documentation" tabindex="-1" onkeydown={trapFocus}>
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
      {#if ask && query.trim()}
        <button type="button" class="ask-row" onclick={askAI} disabled={asking}>
          <span class="kind"><Icon name="bot" size={16} /></span>
          <span class="body"><span class="title">{asking ? 'Thinking…' : `Ask AI: “${query.trim()}”`}</span></span>
        </button>
        {#if askError}<p class="empty">{askError}</p>{/if}
        {#if answer}
          <div class="answer" aria-live="polite">
            {@html miniMarkdown(answer.answer)}
            {#if answer.sources.length}
              <p class="sources">Sources:
                {#each answer.sources as s, i (s.url + i)}<a href={s.url} onclick={() => (open = false)}>{s.title}</a>{/each}
              </p>
            {/if}
          </div>
        {/if}
      {/if}
      {#if !query.trim()}
        <p class="empty">Type to search pages and headings.</p>
      {:else if failed}
        <p class="empty">Search is unavailable right now.</p>
      {:else if !loading && results.length === 0}
        <p class="empty">No results for “{query}”.</p>
      {:else}
        {#if results.length > 0 && results.every((r) => r.fuzzy)}
          <p class="fuzzy-note">No exact matches for “{query}”. Showing similar results.</p>
        {/if}
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
  .fuzzy-note {
    margin: 0.25rem 0.5rem 0.5rem;
    font-size: 0.8rem;
    color: var(--muted-fg);
  }
  .ask-row {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    width: 100%;
    padding: 0.6rem 0.75rem;
    border: 1px dashed var(--border);
    border-radius: 0.6rem;
    background: transparent;
    color: var(--fg);
    font: inherit;
    text-align: left;
    cursor: pointer;
    margin-bottom: 0.25rem;
  }
  .ask-row:hover:not(:disabled) { background: var(--accent); }
  .answer {
    margin: 0.25rem 0 0.75rem;
    padding: 0.75rem 0.9rem;
    border-radius: 0.6rem;
    background: var(--card);
    font-size: 0.9rem;
    line-height: 1.6;
  }
  .answer :global(p) { margin: 0 0 0.6rem; }
  .answer :global(pre) { overflow-x: auto; padding: 0.6rem; border-radius: 0.4rem; background: var(--muted); font-size: 0.8rem; }
  .answer :global(code) { font-family: var(--font-mono); font-size: 0.85em; }
  .answer :global(a) { color: var(--brand); }
  .sources { display: flex; flex-wrap: wrap; gap: 0.5rem; font-size: 0.8rem; color: var(--muted-fg); margin: 0; }
</style>
