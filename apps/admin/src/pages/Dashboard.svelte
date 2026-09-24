<script lang="ts">
  import { api, publicDocsUrl } from '../lib/api.svelte'
  import { href } from '../lib/router.svelte'
  import { formatDate } from '../lib/format'
  import type { Project, Stats } from '../lib/types'

  let stats = $state.raw<Stats | null>(null)
  let projects = $state.raw<Project[]>([])
  let error = $state('')

  Promise.all([api.stats(), api.projects()])
    .then(([s, p]) => {
      stats = s
      projects = p
    })
    .catch((e) => (error = e.message))
</script>

<div class="page-head">
  <div>
    <h1>Dashboard</h1>
    <p>Everything published through jevidocs.</p>
  </div>
  <div class="spacer"></div>
  <a class="btn primary" href={href('/projects')}>Manage projects</a>
</div>

{#if error}<p class="muted">{error}</p>{/if}

<div class="stats">
  {#each [{ label: 'Projects', value: stats?.projects }, { label: 'Pages', value: stats?.pages }, { label: 'API tokens', value: stats?.tokens }] as s (s.label)}
    <div class="card card-pad stat">
      <span class="muted">{s.label}</span>
      <strong>{s.value ?? '—'}</strong>
    </div>
  {/each}
</div>

<h2 class="sub">Projects</h2>
<div class="card">
  {#if projects.length === 0}
    <div class="empty">No projects yet. <a href={href('/projects')}>Create one</a>.</div>
  {:else}
    <table class="list">
      <thead><tr><th>Name</th><th>Visibility</th><th>Updated</th><th></th></tr></thead>
      <tbody>
        {#each projects as p (p.slug)}
          <tr>
            <td>
              <a href={href(`/projects/${p.slug}`)}><strong>{p.name}</strong></a>
              <div class="muted"><code>{p.slug}</code> · {p.description}</div>
            </td>
            <td><span class={['badge', p.public && 'ok']}>{p.public ? 'Public' : 'Private'}</span></td>
            <td class="muted">{formatDate(p.updated_at)}</td>
            <td class="right">
              <a class="btn sm" href={publicDocsUrl(p.slug)} target="_blank" rel="noreferrer">View docs ↗</a>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .stat {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .stat strong {
    font-size: 1.9rem;
    font-weight: 650;
  }
  .sub {
    margin-bottom: 0.75rem;
  }
  a {
    text-decoration: none;
  }
  .right {
    text-align: right;
  }
</style>
