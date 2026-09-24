<script lang="ts">
  import { request } from './api.svelte'
  import { formatDate } from './format'
  import { toast } from './toast.svelte'

  type Source = {
    repo: string
    ref: string
    path: string
    secret: string
    synced_at: string
    status: string
    webhook_url: string
  }

  let { project, onsynced }: { project: string; onsynced?: () => void } = $props()

  const base = $derived(`/api/admin/projects/${encodeURIComponent(project)}/source`)
  let source = $state<Source | null>(null)
  let form = $state({ repo: '', ref: 'main', path: 'docs' })
  let busy = $state(false)
  let syncing = $state(false)

  async function load() {
    try {
      source = await request<Source>('GET', base)
      form = { repo: source.repo, ref: source.ref || 'main', path: source.path }
    } catch (e) {
      toast.error(e)
    }
  }

  $effect(() => {
    void base
    load()
  })

  async function save(e: SubmitEvent) {
    e.preventDefault()
    busy = true
    try {
      source = await request<Source>('PUT', base, form)
      toast.ok(source.repo ? 'GitHub source saved' : 'GitHub source removed')
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  // The sync runs in the background on the API; poll until synced_at moves.
  async function syncNow() {
    if (!source) return
    syncing = true
    const before = source.synced_at
    try {
      await request('POST', `${base}/sync`)
      for (let i = 0; i < 45; i++) {
        await new Promise((r) => setTimeout(r, 2000))
        const s = await request<Source>('GET', base)
        if (s.synced_at !== before) {
          source = s
          if (s.status.startsWith('ok')) {
            toast.ok(`Synced: ${s.status.slice(4)}`)
            onsynced?.()
          } else toast.error(s.status)
          return
        }
      }
      toast.error('Sync is taking long; check the status later')
    } catch (err) {
      toast.error(err)
    } finally {
      syncing = false
    }
  }

  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text)
      toast.ok('Copied')
    } catch {
      toast.error('Copy failed')
    }
  }
</script>

<form class="card card-pad gh" onsubmit={save}>
  <div>
    <h2>GitHub source</h2>
    <p class="muted">
      Sync this project's pages from a folder of Markdown in a public GitHub repository. Pages without a file are
      removed on every sync, so edit the files, not the pages here.
    </p>
  </div>
  <div class="grid">
    <label class="field">
      Repository
      <input type="text" placeholder="owner/repo" bind:value={form.repo} />
    </label>
    <label class="field">
      Branch or tag
      <input type="text" placeholder="main" bind:value={form.ref} />
    </label>
    <label class="field">
      Folder
      <input type="text" placeholder="docs" bind:value={form.path} />
      <span class="hint">Empty for the whole repository.</span>
    </label>
  </div>
  <div class="row">
    <button class="btn primary" type="submit" disabled={busy}>Save source</button>
    <button class="btn" type="button" disabled={syncing || !source?.repo} onclick={syncNow}>
      {syncing ? 'Syncing…' : 'Sync now'}
    </button>
    {#if source?.synced_at}
      <span class="status" class:error={source.status.startsWith('error')}>
        {formatDate(source.synced_at)} · {source.status}
      </span>
    {/if}
  </div>

  {#if source?.repo && source.secret}
    <div class="webhook">
      <h3>Sync on push</h3>
      <p class="muted">
        In the repository go to <em>Settings → Webhooks → Add webhook</em> and use:
      </p>
      <dl>
        <dt>Payload URL</dt>
        <dd><code>{source.webhook_url}</code> <button type="button" class="btn sm" onclick={() => copy(source!.webhook_url)}>Copy</button></dd>
        <dt>Content type</dt>
        <dd><code>application/json</code></dd>
        <dt>Secret</dt>
        <dd><code>{source.secret}</code> <button type="button" class="btn sm" onclick={() => copy(source!.secret)}>Copy</button></dd>
        <dt>Events</dt>
        <dd>Just the push event. Pushes to <code>{source.ref}</code> start a sync.</dd>
      </dl>
    </div>
  {/if}
</form>

<style>
  .gh { display: flex; flex-direction: column; gap: 1rem; margin-top: 1rem; }
  h2 { margin: 0 0 0.25rem; font-size: 1rem; }
  h3 { margin: 0 0 0.25rem; font-size: 0.9rem; }
  .muted { color: var(--muted); margin: 0; font-size: 0.875rem; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr)); gap: 0.75rem; }
  .row { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
  .status { font-size: 0.8rem; color: var(--muted); }
  .status.error { color: var(--danger, #e5484d); }
  .webhook { border-top: 1px solid var(--border); padding-top: 1rem; }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: 0.4rem 1rem; margin: 0.75rem 0 0; font-size: 0.85rem; }
  dt { color: var(--muted); }
  dd { margin: 0; display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
  code { font-size: 0.8rem; word-break: break-all; }
</style>
