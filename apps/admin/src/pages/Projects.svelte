<script lang="ts">
  import { api, publicDocsUrl } from '../lib/api.svelte'
  import { go, href } from '../lib/router.svelte'
  import { formatDate } from '../lib/format'
  import { toast } from '../lib/toast.svelte'
  import ProjectForm from '../lib/ProjectForm.svelte'
  import type { Project, ProjectInput } from '../lib/types'

  let projects = $state.raw<Project[]>([])
  let loading = $state(true)
  let dialog = $state<HTMLDialogElement>()
  let draft = $state<ProjectInput>(blank())
  let busy = $state(false)

  function blank(): ProjectInput {
    return { slug: '', name: '', description: '', github_url: '', links: [], public: true, edit_url: '', banner: '', accent: '', logo_url: '' }
  }

  function load() {
    api
      .projects()
      .then((p) => (projects = p))
      .catch(toast.error)
      .finally(() => (loading = false))
  }
  load()

  function openCreate() {
    draft = blank()
    dialog?.showModal()
  }

  async function create(e: SubmitEvent) {
    e.preventDefault()
    busy = true
    try {
      const p = await api.createProject($state.snapshot(draft))
      dialog?.close()
      toast.ok(`Created ${p.name}`)
      go(`/projects/${p.slug}`)
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }
</script>

<div class="page-head">
  <div>
    <h1>Projects</h1>
    <p>Each project is one documentation site with its own page tree.</p>
  </div>
  <div class="spacer"></div>
  <button class="btn primary" onclick={openCreate}>+ New project</button>
</div>

<div class="grid">
  {#each projects as p (p.slug)}
    <a class="card card-pad project" href={href(`/projects/${p.slug}`)}>
      <div class="row">
        <strong>{p.name}</strong>
        <span class="spacer"></span>
        <span class={['badge', p.public && 'ok']}>{p.public ? 'Public' : 'Private'}</span>
      </div>
      <p class="muted">{p.description || 'No description'}</p>
      <div class="row foot muted">
        <code>{p.slug}</code>
        <span class="spacer"></span>
        <span>{formatDate(p.updated_at)}</span>
      </div>
    </a>
  {:else}
    {#if !loading}
      <div class="card empty wide">No projects yet.</div>
    {/if}
  {/each}
</div>

{#if projects.length}
  <p class="muted note">
    Public docs live at <code>{publicDocsUrl('your-slug')}</code>.
  </p>
{/if}

<dialog bind:this={dialog}>
  <form onsubmit={create}>
    <header><h2>New project</h2></header>
    <div class="body"><ProjectForm bind:value={draft} /></div>
    <footer>
      <button type="button" class="btn" onclick={() => dialog?.close()}>Cancel</button>
      <button type="submit" class="btn primary" disabled={busy}>Create project</button>
    </footer>
  </form>
</dialog>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(18rem, 1fr));
    gap: 1rem;
  }
  .project {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    text-decoration: none;
    transition: border-color 0.15s;
  }
  .project:hover {
    border-color: var(--muted);
  }
  .project p {
    margin: 0;
    flex: 1;
  }
  .foot {
    font-size: 0.8rem;
  }
  .wide {
    grid-column: 1 / -1;
  }
  .note {
    margin-top: 1.5rem;
    font-size: 0.85rem;
  }
  dialog header,
  dialog footer {
    padding: 1rem 1.25rem;
  }
  dialog header {
    border-bottom: 1px solid var(--border);
  }
  dialog .body {
    padding: 1.25rem;
  }
  dialog footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    border-top: 1px solid var(--border);
  }
</style>
