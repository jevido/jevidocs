<script lang="ts">
  import { api, publicDocsUrl } from '../lib/api.svelte'
  import { go, href } from '../lib/router.svelte'
  import { formatDate } from '../lib/format'
  import { toast } from '../lib/toast.svelte'
  import ProjectForm from '../lib/ProjectForm.svelte'
  import FeedbackPanel from '../lib/FeedbackPanel.svelte'
  import AssetsPanel from '../lib/AssetsPanel.svelte'
  import OpenAPIImport from '../lib/OpenAPIImport.svelte'
  import type { AdminPage, ProjectInput } from '../lib/types'

  let { slug }: { slug: string } = $props()

  let form = $state<ProjectInput | null>(null)
  let managed = $state(false)
  let pages = $state.raw<AdminPage[]>([])
  let tab = $state<'pages' | 'assets' | 'feedback' | 'settings'>('pages')
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
  <button role="tab" aria-selected={tab === 'feedback'} onclick={() => (tab = 'feedback')}>Feedback</button>
  <button role="tab" aria-selected={tab === 'settings'} onclick={() => (tab = 'settings')}>Settings</button>
</div>

{#if tab === 'pages'}
  <div class="toolbar">
    <input type="text" placeholder="Filter pages…" bind:value={filter} />
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
          <tr><th>Page</th><th>Section</th><th>Pos.</th><th>Status</th><th>Updated</th><th></th></tr>
        </thead>
        <tbody>
          {#each sorted as p (p.id)}
            <tr class="clickable" onclick={() => go(`/projects/${slug}/pages/${p.id}`)}>
              <td>
                <div class="title" style:padding-left="{Math.max(0, depth(p.slug) - 1) * 1.25}rem">
                  {#if depth(p.slug) > 1}<span class="muted branch">└</span>{/if}
                  <div>
                    <a href={href(`/projects/${slug}/pages/${p.id}`)}>{p.title || 'Untitled'}</a>
                    <div class="muted"><code>/{p.slug}</code></div>
                  </div>
                </div>
              </td>
              <td class="muted">{p.section || ''}</td>
              <td class="muted">{p.position}</td>
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
{:else if tab === 'feedback'}
  <FeedbackPanel project={slug} />
{:else if form}
  <form class="card card-pad settings" onsubmit={save}>
    <ProjectForm bind:value={form} lockSlug />
    <div class="row">
      <button class="btn primary" type="submit" disabled={busy}>Save changes</button>
    </div>
  </form>

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
</style>
