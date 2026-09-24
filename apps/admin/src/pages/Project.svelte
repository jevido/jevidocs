<script lang="ts">
  import { api, publicDocsUrl } from '../lib/api.svelte'
  import { go, href } from '../lib/router.svelte'
  import { formatDate } from '../lib/format'
  import { toast } from '../lib/toast.svelte'
  import ProjectForm from '../lib/ProjectForm.svelte'
  import FeedbackPanel from '../lib/FeedbackPanel.svelte'
  import InsightsPanel from '../lib/InsightsPanel.svelte'
  import AssetsPanel from '../lib/AssetsPanel.svelte'
  import OpenAPIImport from '../lib/OpenAPIImport.svelte'
  import GitHubSource from '../lib/GitHubSource.svelte'
  import type { AdminPage, ProjectInput } from '../lib/types'

  let { slug }: { slug: string } = $props()

  let form = $state<ProjectInput | null>(null)
  let managed = $state(false)
  let pages = $state.raw<AdminPage[]>([])
  let tab = $state<'pages' | 'assets' | 'insights' | 'feedback' | 'settings'>('pages')
  let busy = $state(false)
  let filter = $state('')

  function load() {
    api
      .projects()
      .then((all) => {
        const p = all.find((x) => x.slug === slug)
        if (!p) throw new Error(`Project "${slug}" not found`)
        form = {
          slug: p.slug,
          name: p.name,
          description: p.description,
          github_url: p.github_url,
          links: (p.links ?? []).map((l) => ({ ...l })),
          public: p.public,
          edit_url: p.edit_url ?? '',
          banner: p.banner ?? '',
          accent: p.accent ?? '',
          logo_url: p.logo_url ?? '',
          version_group: p.version_group ?? '',
          version_label: p.version_label ?? '',
          locales: (p.locales ?? []).length > 1 ? (p.locales ?? []).join(',') : '',
          default_locale: p.default_locale ?? 'en',
        }
        managed = !!p.managed
      })
      .catch(toast.error)
    api
      .pages(slug)
      .then((p) => (pages = p))
      .catch(toast.error)
  }
  load()

  // Tree order: parents before children, siblings by position then title.
  const sorted = $derived.by(() => {
    const q = filter.trim().toLowerCase()
    const list = q
      ? pages.filter((p) => p.slug.toLowerCase().includes(q) || p.title.toLowerCase().includes(q))
      : pages
    const key = (p: AdminPage) => p.slug.split('/').filter(Boolean)
    const byPath = new Map(pages.map((p) => [p.slug, p]))
    const pos = (path: string[]) => byPath.get(path.join('/'))?.position ?? 0
    return [...list].sort((a, b) => {
      const ka = key(a)
      const kb = key(b)
      for (let i = 0; i < Math.min(ka.length, kb.length); i++) {
        if (ka[i] === kb[i]) continue
        const pa = pos(ka.slice(0, i + 1))
        const pb = pos(kb.slice(0, i + 1))
        return pa - pb || ka[i].localeCompare(kb[i])
      }
      return ka.length - kb.length
    })
  })

  const depth = (s: string) => s.split('/').filter(Boolean).length

  // Reordering: pages move among siblings (same parent slug). Positions of
  // the whole sibling group are renumbered 10, 20, 30… and saved at once.
  const parentOf = (s: string) => s.split('/').filter(Boolean).slice(0, -1).join('/')
  let dragId = $state<number | null>(null)
  let dropId = $state<number | null>(null)
  let reordering = $state(false)

  function siblingsOf(p: AdminPage): AdminPage[] {
    const parent = parentOf(p.slug)
    return sorted.filter((x) => parentOf(x.slug) === parent)
  }

  async function saveOrder(group: AdminPage[]) {
    const items = group.map((p, i) => ({ id: p.id, position: (i + 1) * 10 }))
    const byId = new Map(items.map((it) => [it.id, it.position]))
    const before = pages
    pages = pages.map((p) => (byId.has(p.id) ? { ...p, position: byId.get(p.id)! } : p))
    reordering = true
    try {
      await api.reorder(slug, items)
      toast.ok('Order saved')
    } catch (e) {
      pages = before
      toast.error(e)
    } finally {
      reordering = false
    }
  }

  function move(p: AdminPage, delta: number, e: MouseEvent) {
    e.stopPropagation()
    const group = siblingsOf(p)
    const i = group.findIndex((x) => x.id === p.id)
    const j = i + delta
    if (j < 0 || j >= group.length) return
    ;[group[i], group[j]] = [group[j], group[i]]
    saveOrder(group)
  }

  function canDrop(target: AdminPage): boolean {
    const dragged = pages.find((x) => x.id === dragId)
    return !!dragged && dragged.id !== target.id && parentOf(dragged.slug) === parentOf(target.slug)
  }

  function drop(target: AdminPage) {
    const dragged = pages.find((x) => x.id === dragId)
    dragId = dropId = null
    if (!dragged || !canDropPair(dragged, target)) return
    const group = siblingsOf(target)
    const from = group.findIndex((x) => x.id === dragged.id)
    const to = group.findIndex((x) => x.id === target.id)
    group.splice(from, 1)
    group.splice(to, 0, dragged)
    saveOrder(group)
  }

  const canDropPair = (a: AdminPage, b: AdminPage) => a.id !== b.id && parentOf(a.slug) === parentOf(b.slug)

  async function save(e: SubmitEvent) {
    e.preventDefault()
    if (!form) return
    busy = true
    try {
      await api.updateProject(slug, $state.snapshot(form))
      toast.ok('Project saved')
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  async function remove() {
    if (!confirm(`Delete project "${slug}" and all of its pages? This cannot be undone.`)) return
    try {
      await api.deleteProject(slug)
      toast.ok('Project deleted')
      go('/projects')
    } catch (err) {
      toast.error(err)
    }
  }

  async function removePage(p: AdminPage, e: MouseEvent) {
    e.stopPropagation()
    if (!confirm(`Delete page "${p.title}" (/${p.slug})?`)) return
    try {
      await api.deletePage(slug, p.id)
      pages = pages.filter((x) => x.id !== p.id)
      toast.ok('Page deleted')
    } catch (err) {
      toast.error(err)
    }
  }
</script>

<div class="crumbs muted"><a href={href('/projects')}>Projects</a> / {slug}</div>
<div class="page-head">
  <div>
    <h1>{form?.name ?? slug}</h1>
    {#if form?.description}<p>{form.description}</p>{/if}
  </div>
  <div class="spacer"></div>
  <OpenAPIImport project={slug} onimported={load} />
  <a class="btn" href={publicDocsUrl(slug)} target="_blank" rel="noreferrer">View docs ↗</a>
  <a class="btn primary" href={href(`/projects/${slug}/pages/new`)}>+ New page</a>
</div>

{#if managed}
  <p class="managed-note" role="note">
    This project is synced from files in the repository on every API start. Edits made here are overwritten by
    the next deploy; change <code>services/api/content/docs</code> instead.
  </p>
{/if}

<div class="tabs" role="tablist">
  <button role="tab" aria-selected={tab === 'pages'} onclick={() => (tab = 'pages')}>Pages <span class="badge">{pages.length}</span></button>
  <button role="tab" aria-selected={tab === 'assets'} onclick={() => (tab = 'assets')}>Assets</button>
  <button role="tab" aria-selected={tab === 'insights'} onclick={() => (tab = 'insights')}>Insights</button>
  <button role="tab" aria-selected={tab === 'feedback'} onclick={() => (tab = 'feedback')}>Feedback</button>
  <button role="tab" aria-selected={tab === 'settings'} onclick={() => (tab = 'settings')}>Settings</button>
</div>

{#if tab === 'pages'}
  <div class="toolbar">
    <input type="text" placeholder="Filter pages…" bind:value={filter} />
    <span class="muted hint-order">{reordering ? 'Saving order…' : filter ? '' : 'Drag rows or use ↑/↓ to reorder siblings.'}</span>
  </div>
  <div class="card">
    {#if sorted.length === 0}
      <div class="empty">
        {pages.length ? 'No pages match.' : 'No pages yet.'}
        <a href={href(`/projects/${slug}/pages/new`)}>Write the first one</a>.
      </div>
    {:else}
      <table class="list">
        <thead>
          <tr><th>Page</th><th>Section</th><th>Pos.</th><th>Status</th><th>Updated</th><th><span class="sr-only">Actions</span></th></tr>
        </thead>
        <tbody>
          {#each sorted as p (p.id)}
            <tr
              class="clickable"
              class:dragging={dragId === p.id}
              class:drop-target={dropId === p.id}
              draggable={!filter && !reordering}
              ondragstart={(e) => {
                dragId = p.id
                e.dataTransfer?.setData('text/plain', String(p.id))
              }}
              ondragover={(e) => {
                if (canDrop(p)) {
                  e.preventDefault()
                  dropId = p.id
                }
              }}
              ondragleave={() => dropId === p.id && (dropId = null)}
              ondrop={(e) => {
                e.preventDefault()
                drop(p)
              }}
              ondragend={() => (dragId = dropId = null)}
              onclick={() => go(`/projects/${slug}/pages/${p.id}`)}>
              <td>
                <div class="title" style:padding-left="{Math.max(0, depth(p.slug) - 1) * 1.25}rem">
                  {#if depth(p.slug) > 1}<span class="muted branch">└</span>{/if}
                  <div>
                    <a href={href(`/projects/${slug}/pages/${p.id}`)}>{p.title || 'Untitled'}</a>
                    <div class="muted"><code>/{p.slug}</code>{#if p.locale}&nbsp;<span class="badge">{p.locale}</span>{/if}</div>
                  </div>
                </div>
              </td>
              <td class="muted">{p.section || ''}</td>
              <td class="muted nowrap">
                {#if !filter}
                  <span class="order">
                    <button class="btn sm ghost" aria-label="Move {p.title} up" disabled={reordering} onclick={(e) => move(p, -1, e)}>↑</button>
                    <button class="btn sm ghost" aria-label="Move {p.title} down" disabled={reordering} onclick={(e) => move(p, 1, e)}>↓</button>
                  </span>
                {/if}
                {p.position}
              </td>
              <td>
                <span class={['badge', p.published ? 'ok' : 'warn']}>{p.published ? 'Published' : 'Draft'}</span>
              </td>
              <td class="muted nowrap">{formatDate(p.updated_at)}</td>
              <td class="right">
                <button class="btn sm ghost danger" onclick={(e) => removePage(p, e)}>Delete</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
{:else if tab === 'assets'}
  <AssetsPanel project={slug} />
{:else if tab === 'insights'}
  <InsightsPanel project={slug} {pages} />
{:else if tab === 'feedback'}
  <FeedbackPanel project={slug} />
{:else if form}
  <form class="card card-pad settings" onsubmit={save}>
    <ProjectForm bind:value={form} lockSlug />
    <div class="row">
      <button class="btn primary" type="submit" disabled={busy}>Save changes</button>
    </div>
  </form>

  <GitHubSource project={slug} onsynced={load} />

  <div class="card card-pad danger-zone">
    <div>
      <h2>Delete project</h2>
      <p class="muted">Removes the project and every page in it.</p>
    </div>
    <button class="btn danger" onclick={remove}>Delete project</button>
  </div>
{/if}

<style>
  .crumbs {
    font-size: 0.85rem;
    margin-bottom: 0.5rem;
  }
  .crumbs a {
    text-decoration: none;
  }
  .tabs {
    display: flex;
    gap: 0.25rem;
    border-bottom: 1px solid var(--border);
    margin-bottom: 1rem;
  }
  .tabs button {
    font: inherit;
    font-weight: 500;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--muted);
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    display: flex;
    gap: 0.4rem;
    align-items: center;
    margin-bottom: -1px;
  }
  .tabs button[aria-selected='true'] {
    color: var(--text);
    border-bottom-color: var(--text);
  }
  .toolbar {
    margin-bottom: 0.75rem;
    max-width: 20rem;
  }
  .title {
    display: flex;
    gap: 0.5rem;
    align-items: flex-start;
  }
  .title a {
    font-weight: 500;
    text-decoration: none;
  }
  .branch {
    font-family: var(--mono);
  }
  .right {
    text-align: right;
  }
  .nowrap {
    white-space: nowrap;
  }
  .settings {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    max-width: 48rem;
  }
  .danger-zone {
    margin-top: 1.5rem;
    max-width: 48rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
  }
  .danger-zone div {
    flex: 1;
  }
  .danger-zone p {
    margin: 0.25rem 0 0;
  }
  .managed-note {
    margin: 0 0 1rem;
    padding: 0.6rem 0.8rem;
    border: 1px solid color-mix(in oklab, orange 40%, var(--border));
    background: color-mix(in oklab, orange 10%, transparent);
    border-radius: 0.5rem;
    font-size: 0.85rem;
  }
  tr.dragging { opacity: 0.45; }
  tr.drop-target td { box-shadow: inset 0 2px 0 var(--accent, #6366f1); }
  tr[draggable='true'] { cursor: grab; }
  .order { display: inline-flex; gap: 0.1rem; margin-right: 0.35rem; }
  .order .btn { padding: 0 0.3rem; min-width: 0; }
  .hint-order { font-size: 0.8rem; margin-left: 0.75rem; }
</style>
