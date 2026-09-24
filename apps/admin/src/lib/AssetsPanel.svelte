<script lang="ts">
  import { assetsApi, assetMarkdown, formatSize, ASSET_ACCEPT, type Asset } from './assets-api'
  import { toast } from './toast.svelte'

  let { project }: { project: string } = $props()

  let assets = $state.raw<Asset[]>([])
  let loading = $state(true)
  let uploading = $state(0)
  let dragging = $state(false)

  function load() {
    assetsApi
      .list(project)
      .then((a) => (assets = a))
      .catch(toast.error)
      .finally(() => (loading = false))
  }
  load()

  async function uploadFiles(files: FileList | File[] | null) {
    if (!files) return
    for (const f of Array.from(files)) {
      uploading++
      try {
        await assetsApi.upload(project, f)
        toast.ok(`Uploaded ${f.name}`)
      } catch (e) {
        toast.error(e)
      } finally {
        uploading--
      }
    }
    load()
  }

  async function copy(a: Asset) {
    try {
      await navigator.clipboard.writeText(assetMarkdown(a))
      toast.ok('Markdown copied')
    } catch {
      toast.error('Copy failed')
    }
  }

  async function remove(a: Asset) {
    if (!confirm(`Delete ${a.name}? Pages that embed it will show a broken image.`)) return
    try {
      await assetsApi.remove(project, a.id)
      assets = assets.filter((x) => x.id !== a.id)
      toast.ok('Asset deleted')
    } catch (e) {
      toast.error(e)
    }
  }

  function ondrop(e: DragEvent) {
    e.preventDefault()
    dragging = false
    uploadFiles(e.dataTransfer?.files ?? null)
  }
</script>

<div
  class="drop card card-pad"
  class:dragging
  role="region"
  aria-label="Upload assets"
  ondragover={(e) => {
    e.preventDefault()
    dragging = true
  }}
  ondragleave={() => (dragging = false)}
  {ondrop}
>
  <p>Drop images or files here, or</p>
  <label class="btn">
    Choose files
    <input
      type="file"
      multiple
      accept={ASSET_ACCEPT}
      hidden
      onchange={(e) => uploadFiles((e.currentTarget as HTMLInputElement).files)}
    />
  </label>
  <p class="muted small">PNG, JPEG, GIF, WebP, SVG, AVIF, PDF or text, up to 8 MB. {uploading ? `Uploading ${uploading}…` : ''}</p>
</div>

{#if assets.length === 0}
  {#if !loading}<div class="empty">No assets yet.</div>{/if}
{:else}
  <div class="grid">
    {#each assets as a (a.id)}
      <div class="card asset">
        <a class="thumb" href={a.url} target="_blank" rel="noreferrer">
          {#if a.content_type.startsWith('image/')}
            <img src={a.url} alt={a.name} loading="lazy" />
          {:else}
            <span class="ext">{a.name.split('.').pop()?.toUpperCase()}</span>
          {/if}
        </a>
        <div class="meta">
          <strong title={a.name}>{a.name}</strong>
          <span class="muted small">{formatSize(a.size)}</span>
        </div>
        <div class="actions">
          <button class="btn sm" onclick={() => copy(a)}>Copy Markdown</button>
          <button class="btn sm ghost danger" onclick={() => remove(a)}>Delete</button>
        </div>
      </div>
    {/each}
  </div>
{/if}

<style>
  .drop {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.75rem;
    border-style: dashed;
    margin-bottom: 1rem;
  }
  .drop p { margin: 0; }
  .drop.dragging { background: var(--surface-2); }
  .small { font-size: 0.8rem; width: 100%; }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: 0.75rem;
  }
  .asset { display: flex; flex-direction: column; overflow: hidden; }
  .thumb {
    display: grid;
    place-items: center;
    height: 8rem;
    background: var(--surface-2);
    border-bottom: 1px solid var(--border);
  }
  .thumb img { max-width: 100%; max-height: 100%; object-fit: contain; }
  .ext { font-weight: 700; color: var(--muted); }
  .meta { display: flex; flex-direction: column; padding: 0.5rem 0.75rem 0; min-width: 0; }
  .meta strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.85rem; }
  .actions { display: flex; gap: 0.4rem; padding: 0.5rem 0.75rem 0.75rem; }
</style>
