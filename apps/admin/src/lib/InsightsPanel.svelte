<script lang="ts">
  import { insightsApi, type Insights } from './insights-api'
  import { href } from './router.svelte'
  import { toast } from './toast.svelte'
  import type { AdminPage } from './types'

  let { project, pages }: { project: string; pages: AdminPage[] } = $props()

  let days = $state(30)
  let data = $state.raw<Insights | null>(null)

  $effect(() => {
    const d = days
    data = null
    insightsApi
      .insights(project, d)
      .then((r) => (data = r))
      .catch(toast.error)
  })

  const idBySlug = $derived(new Map(pages.map((p) => [p.slug, p.id])))

  // Chart geometry: one bar per day in a 600x140 viewBox.
  const W = 600
  const H = 140
  const max = $derived(Math.max(1, ...(data?.views_by_day ?? []).map((d) => d.views)))
  const barW = $derived(data ? W / Math.max(1, data.views_by_day.length) : 0)
  const score = $derived(
    data && data.feedback.helpful + data.feedback.not_helpful > 0
      ? Math.round((data.feedback.helpful / (data.feedback.helpful + data.feedback.not_helpful)) * 100) + '%'
      : '—',
  )
</script>

<div class="toolbar">
  <label>
    Period
    <select bind:value={days}>
      <option value={7}>Last 7 days</option>
      <option value={30}>Last 30 days</option>
      <option value={90}>Last 90 days</option>
    </select>
  </label>
</div>

{#if !data}
  <p class="muted">Loading…</p>
{:else}
  <div class="kpis">
    <div class="card card-pad"><span class="muted">Page views</span><strong>{data.views_total}</strong></div>
    <div class="card card-pad"><span class="muted">Searches</span><strong>{[...data.top_searches, ...data.zero_result_searches].reduce((n, s) => n + s.count, 0)}</strong></div>
    <div class="card card-pad"><span class="muted">Helpful</span><strong>{score}</strong></div>
  </div>

  <div class="card card-pad">
    <h2>Views per day</h2>
    <svg viewBox="0 0 {W} {H + 18}" class="chart" role="img" aria-label="Page views per day">
      <line x1="0" y1={H} x2={W} y2={H} class="axis" />
      {#each data.views_by_day as d, i (d.day)}
        {@const h = (d.views / max) * (H - 8)}
        <rect x={i * barW + barW * 0.15} y={H - h} width={barW * 0.7} height={Math.max(h, d.views ? 1 : 0)} rx="2">
          <title>{d.day}: {d.views} views</title>
        </rect>
      {/each}
      <text x="0" y={H + 14}>{data.views_by_day[0]?.day}</text>
      <text x={W} y={H + 14} text-anchor="end">{data.views_by_day.at(-1)?.day}</text>
      <text x={W} y="10" text-anchor="end">max {max}</text>
    </svg>
  </div>

  <div class="grid">
    <div class="card card-pad">
      <h2>Top pages</h2>
      {#if data.top_pages.length === 0}
        <p class="muted">No views yet. The docs site counts one view per page load.</p>
      {:else}
        <table class="list">
          <thead><tr><th>Page</th><th>Views</th></tr></thead>
          <tbody>
            {#each data.top_pages as p (p.slug)}
              {@const id = idBySlug.get(p.slug)}
              <tr>
                <td>
                  {#if id}<a href={href(`/projects/${project}/pages/${id}`)}>{p.title}</a>{:else}{p.title}{/if}
                  <div class="muted"><code>/{p.slug}</code></div>
                </td>
                <td>{p.views}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>

    <div class="card card-pad">
      <h2>Top searches</h2>
      {#if data.top_searches.length === 0}
        <p class="muted">No searches yet.</p>
      {:else}
        <ul class="terms">
          {#each data.top_searches as s (s.query)}
            <li><span>{s.query}</span><span class="muted">{s.count}× · {s.results} results</span></li>
          {/each}
        </ul>
      {/if}
      <h2 class="gap-title">Content gaps</h2>
      <p class="muted small">Searches that found nothing: pages worth writing.</p>
      {#if data.zero_result_searches.length === 0}
        <p class="muted">None. Every search found something.</p>
      {:else}
        <ul class="terms">
          {#each data.zero_result_searches as s (s.query)}
            <li><span>{s.query}</span><span class="muted">{s.count}×</span></li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>
{/if}

<style>
  .toolbar { display: flex; justify-content: flex-end; margin-bottom: 1rem; }
  .toolbar label { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; }
  .kpis { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; margin-bottom: 1rem; }
  .kpis .card { display: flex; flex-direction: column; gap: 0.25rem; }
  .kpis strong { font-size: 1.6rem; }
  h2 { font-size: 1rem; margin: 0 0 0.75rem; }
  .chart { width: 100%; height: auto; display: block; }
  .chart rect { fill: var(--accent); }
  .chart .axis { stroke: var(--border); }
  .chart text { font-size: 10px; fill: var(--muted); }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(20rem, 1fr)); gap: 1rem; margin-top: 1rem; }
  .terms { list-style: none; margin: 0; padding: 0; }
  .terms li { display: flex; justify-content: space-between; gap: 1rem; padding: 0.4rem 0; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
  .gap-title { margin-top: 1.5rem; margin-bottom: 0.25rem; }
  .small { font-size: 0.8rem; margin: 0 0 0.5rem; }
  @media (max-width: 640px) { .kpis { grid-template-columns: 1fr; } }
</style>
