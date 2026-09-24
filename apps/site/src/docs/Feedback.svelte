<script lang="ts">
  import { API_URL } from '../lib/api'

  let { project, slug }: { project: string; slug: string } = $props()

  type Vote = 'up' | 'down'
  const key = $derived(`jevidocs:feedback:${project}:${slug}`)

  let vote = $state<Vote | null>(null)
  let message = $state('')
  let phase = $state<'ask' | 'details' | 'sending' | 'done'>('ask')
  let error = $state('')

  // Pages remember the reader's earlier answer so they are not asked twice.
  $effect(() => {
    let saved: string | null = null
    try {
      saved = localStorage.getItem(key)
    } catch {
      saved = null
    }
    const remembered: Vote | null = saved === 'up' || saved === 'down' ? saved : null
    vote = remembered
    phase = remembered ? 'done' : 'ask'
    message = ''
    error = ''
  })

  function choose(v: Vote) {
    vote = v
    phase = 'details'
  }

  async function send() {
    if (!vote) return
    phase = 'sending'
    error = ''
    try {
      const res = await fetch(`${API_URL}/api/projects/${encodeURIComponent(project)}/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({ slug, helpful: vote === 'up', message: message.trim() }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => null)
        throw new Error(body?.error ?? `HTTP ${res.status}`)
      }
      try {
        localStorage.setItem(key, vote)
      } catch {
        // Private mode: the vote is still sent, just not remembered.
      }
      phase = 'done'
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not send feedback'
      phase = 'details'
    }
  }
</script>

<section class="feedback" aria-label="Page feedback">
  {#if phase === 'done'}
    <p class="thanks">Thank you for your feedback!</p>
  {:else}
    <div class="row">
      <p class="q">Was this page helpful?</p>
      <div class="votes">
        <button type="button" class:active={vote === 'up'} aria-pressed={vote === 'up'} onclick={() => choose('up')}>
          <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M7 10v12" /><path d="M15 5.88 14 10h5.83a2 2 0 0 1 1.92 2.56l-2.33 8A2 2 0 0 1 17.5 22H4a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h2.76a2 2 0 0 0 1.79-1.11L12 2a3.13 3.13 0 0 1 3 3.88Z" /></svg>
          Good
        </button>
        <button type="button" class:active={vote === 'down'} aria-pressed={vote === 'down'} onclick={() => choose('down')}>
          <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17 14V2" /><path d="M9 18.12 10 14H4.17a2 2 0 0 1-1.92-2.56l2.33-8A2 2 0 0 1 6.5 2H20a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2h-2.76a2 2 0 0 0-1.79 1.11L12 22a3.13 3.13 0 0 1-3-3.88Z" /></svg>
          Bad
        </button>
      </div>
    </div>
    {#if phase === 'details' || phase === 'sending'}
      <form
        class="details"
        onsubmit={(e) => {
          e.preventDefault()
          send()
        }}>
        <textarea bind:value={message} maxlength="2000" rows="3" placeholder="Tell us more (optional)" aria-label="Tell us more (optional)"></textarea>
        {#if error}<p class="error">{error}</p>{/if}
        <button type="submit" class="send" disabled={phase === 'sending'}>{phase === 'sending' ? 'Sending…' : 'Send'}</button>
      </form>
    {/if}
  {/if}
</section>

<style>
  .feedback {
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    padding: 0.85rem 1rem;
    margin: 0 0 1.25rem;
  }
  .row { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; flex-wrap: wrap; }
  .q, .thanks { margin: 0; font-size: 0.9rem; font-weight: 500; }
  .thanks { color: var(--muted-fg); font-weight: 400; }
  .votes { display: flex; gap: 0.4rem; }
  button {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font: inherit;
    font-size: 0.85rem;
    padding: 0.35rem 0.75rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted-fg);
    cursor: pointer;
  }
  button:hover { color: var(--fg); background: var(--accent); }
  button.active { color: var(--fg); background: var(--accent); border-color: color-mix(in oklab, var(--brand) 40%, var(--border)); }
  .details { display: flex; flex-direction: column; gap: 0.5rem; margin-top: 0.75rem; }
  textarea {
    font: inherit;
    font-size: 0.9rem;
    resize: vertical;
    padding: 0.6rem 0.7rem;
    border-radius: 0.5rem;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg);
  }
  textarea:focus { outline: 2px solid color-mix(in oklab, var(--brand) 50%, transparent); outline-offset: 1px; }
  .send { align-self: flex-end; color: var(--fg); }
  .send:disabled { opacity: 0.6; cursor: default; }
  .error { margin: 0; font-size: 0.8rem; color: #dc2626; }
</style>
