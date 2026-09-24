<script lang="ts">
  import { extraApi, lineDiff, type Revision, type RevisionSummary } from './extra-api'
  import { formatDate } from './format'
  import { toast } from './toast.svelte'
  import type { AdminPage } from './types'

  let {
    project,
    pageId,
    current,
    dirty,
    onclose,
    onrestore,
  }: {
    project: string
    pageId: number
    current: string
    dirty: boolean
    onclose: () => void
    onrestore: (page: AdminPage) => void
  } = $props()

  let revisions = $state.raw<RevisionSummary[]>([])
  let loading = $state(true)
  let selected = $state<Revision | null>(null)
  let mode = $state<'diff' | 'source'>('diff')
  let busy = $state(false)

  function load() {
    loading = true
    extraApi
      .revisions(project, pageId)
      .then((r) => (revisions = r))
      .catch(toast.error)
      .finally(() => (loading = false))
  }
  load()

  async function select(id: number) {
    try {
      selected = await extraApi.revision(project, pageId, id)
    } catch (e) {
      toast.error(e)
    }
  }

  async function restore() {
    if (!selected || busy) return
    const warn = dirty ? '\n\nYour unsaved changes in the editor will be lost.' : ''
    if (!confirm(`Restore the version from ${formatDate(selected.created_at)}? The current version is kept in the history.${warn}`)) return
    busy = true
    try {
      const page = await extraApi.restore(project, pageId, selected.id)
      toast.ok('Version restored')
      onrestore(page)
      selected = null
      load()
    } catch (e) {
      toast.error(e)
    } finally {
      busy = false
    }
  }

  const diff = $derived(selected ? lineDiff(selected.body, current) : [])
  const changes = $derived(diff.filter((l) => l.kind !== 'same').length)
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && onclose()} />

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="backdrop" onclick={onclose}></div>
<aside class="drawer" aria-label="Page history">
  <header>
    <h2>History</h2>
    <button class="btn" onclick={onclose} aria-label="Close history">✕</button>
  </header>

  {#if selected}
    <div class="sel-head">
      <button class="btn" onclick={() => (selected = null)}>← All versions</button>
      <div class="spacer"></div>
      <button class="btn primary" onclick={restore} disabled={busy}>{busy ? 'Restoring…' : 'Restore this version'}</button>
    </div>
    <p class="muted small">
      {selected.title} · {formatDate(selected.created_at)} · {selected.size} bytes
    </p>
    <div class="modes" role="tablist">
      <button role="tab" aria-selected={mode === 'diff'} onclick={() => (mode = 'diff')}>Changes since ({changes})</button>
      <button role="tab" aria-selected={mode === 'source'} onclick={() => (mode = 'source')}>Markdown</button>
    </div>
    {#if mode === 'diff'}
      <p class="muted small">Red lines are only in this version, green lines only in the editor.</p>
      <pre class="diff">{#each diff as l, i (i)}<span class={l.kind}>{l.kind === 'add' ? '+ ' : l.kind === 'del' ? '- ' : '  '}{l.text}
</span>{/each}</pre>
    {:else}
      <pre class="diff">{selected.body}</pre>
    {/if}
  {:else if loading}
    <p class="muted">Loading…</p>
  {:else if revisions.length === 0}
    <p class="muted">No earlier versions yet. Every save that changes the title, description or body keeps the previous version here (up to 50).</p>
  {:else}
    <ul class="revs">
      {#each revisions as r (r.id)}
        <li>
          <button onclick={() => select(r.id)}>
            <span class="when">{formatDate(r.created_at)}</span>
            <span class="muted small">{r.title} · {r.size} bytes</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</aside>

<style>
  .backdrop { position: fixed; inset: 0; background: rgb(0 0 0 / 0.35); z-index: 40; }
  .drawer {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: min(44rem, 100vw);
    background: var(--bg);
    border-left: 1px solid var(--border);
    z-index: 41;
    padding: 1rem 1.25rem;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }
  header, .sel-head { display: flex; align-items: center; gap: 0.5rem; }
  header { justify-content: space-between; }
  h2 { margin: 0; font-size: 1.1rem; }
  .spacer { flex: 1; }
  .small { font-size: 0.8rem; margin: 0; }
  .revs { list-style: none; margin: 0; padding: 0; }
  .revs button {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.15rem;
    padding: 0.6rem 0.7rem;
    border: 0;
    border-bottom: 1px solid var(--border);
    background: transparent;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  .revs button:hover { background: var(--surface-2, #8881); }
  .when { font-weight: 500; }
  .modes { display: flex; gap: 0.25rem; }
  .modes button {
    font: inherit;
    font-size: 0.85rem;
    padding: 0.3rem 0.7rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    cursor: pointer;
  }
  .modes button[aria-selected='true'] { color: inherit; background: var(--surface-2, #8882); }
  .diff {
    margin: 0;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    font-size: 0.8rem;
    line-height: 1.5;
    overflow: auto;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .diff .add { background: rgb(22 163 74 / 0.15); color: #16a34a; display: block; }
  .diff .del { background: rgb(220 38 38 / 0.12); color: #dc2626; display: block; }
  .diff .same { display: block; }
</style>
