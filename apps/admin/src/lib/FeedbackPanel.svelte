<script lang="ts">
  import { extraApi, type FeedbackEntry, type FeedbackTotal } from './extra-api'
  import { formatDate } from './format'
  import { toast } from './toast.svelte'

  let { project }: { project: string } = $props()

  let entries = $state.raw<FeedbackEntry[]>([])
  let totals = $state.raw<FeedbackTotal[]>([])
  let loaded = $state(false)
  let onlyMessages = $state(false)

  $effect(() => {
    loaded = false
    extraApi
      .feedback(project)
      .then((r) => {
        entries = r.entries
        totals = r.totals
      })
      .catch(toast.error)
      .finally(() => (loaded = true))
  })

  const shown = $derived(onlyMessages ? entries.filter((e) => e.message) : entries)
  const score = (t: FeedbackTotal) => Math.round((t.helpful / Math.max(1, t.helpful + t.not_helpful)) * 100)
</script>

<div class="card card-pad">
  <h2>Per page</h2>
  {#if !loaded}
    <p class="muted">Loading…</p>
  {:else if totals.length === 0}
    <p class="muted">No feedback yet. Readers answer "Was this page helpful?" at the bottom of every page.</p>
  {:else}
    <table class="list">
      <thead><tr><th>Page</th><th>Helpful</th><th>Not helpful</th><th>Score</th></tr></thead>
      <tbody>
        {#each totals as t (t.slug)}
          <tr>
            <td><code>{t.slug || '(index)'}</code></td>
            <td>{t.helpful}</td>
            <td>{t.not_helpful}</td>
            <td>
              <span class="bar"><span style:width="{score(t)}%"></span></span>
              {score(t)}%
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

{#if entries.length}
  <div class="card card-pad">
    <div class="head">
      <h2>Answers</h2>
      <label><input type="checkbox" bind:checked={onlyMessages} /> Only with a message</label>
    </div>
    <ul class="entries">
      {#each shown as e (e.id)}
        <li>
          <span class="vote" class:up={e.helpful}>{e.helpful ? '👍' : '👎'}</span>
          <div>
            <div class="meta"><code>{e.slug || '(index)'}</code> · {formatDate(e.created_at)}</div>
            {#if e.message}<p>{e.message}</p>{/if}
          </div>
        </li>
      {/each}
    </ul>
  </div>
{/if}

<style>
  h2 { font-size: 1rem; margin: 0 0 0.75rem; }
  .card + .card { margin-top: 1rem; }
  table { width: 100%; }
  .bar { display: inline-block; width: 5rem; height: 0.4rem; border-radius: 999px; background: var(--surface-2, #8883); vertical-align: middle; margin-right: 0.4rem; overflow: hidden; }
  .bar span { display: block; height: 100%; background: #16a34a; }
  .head { display: flex; justify-content: space-between; align-items: center; }
  .head label { font-size: 0.85rem; display: flex; gap: 0.35rem; align-items: center; }
  .entries { list-style: none; margin: 0; padding: 0; }
  .entries li { display: flex; gap: 0.75rem; padding: 0.6rem 0; border-top: 1px solid var(--border); }
  .entries li:first-child { border-top: 0; }
  .meta { font-size: 0.8rem; color: var(--muted); }
  .entries p { margin: 0.25rem 0 0; white-space: pre-wrap; }
</style>
