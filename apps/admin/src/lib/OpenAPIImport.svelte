<script lang="ts">
  import { api } from './api.svelte'
  import { toast } from './toast.svelte'

  let { project, onimported }: { project: string; onimported?: () => void } = $props()

  let dialog = $state<HTMLDialogElement>()
  let spec = $state('')
  let prefix = $state('api-reference')
  let busy = $state(false)

  async function pick(e: Event) {
    const file = (e.currentTarget as HTMLInputElement).files?.[0]
    if (file) spec = await file.text()
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    if (!spec.trim()) return toast.error('Paste or choose an OpenAPI document first')
    busy = true
    try {
      const r = await api.importOpenAPI(project, spec, prefix)
      toast.ok(`OpenAPI imported: ${r.created} created, ${r.updated} updated, ${r.unchanged} unchanged, ${r.deleted} deleted`)
      dialog?.close()
      onimported?.()
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }
</script>

<button class="btn" type="button" onclick={() => dialog?.showModal()}>Import OpenAPI</button>

<dialog bind:this={dialog}>
  <form onsubmit={submit}>
    <header><h2>Import OpenAPI</h2></header>
    <div class="body">
      <p class="muted">
        Generates an API reference (a page per operation, grouped by tag) from an OpenAPI 3 document in JSON or
        YAML. Re-importing updates the pages and removes ones the spec no longer has.
      </p>
      <label>
        <span>Prefix</span>
        <input type="text" bind:value={prefix} placeholder="api-reference" />
      </label>
      <label>
        <span>File</span>
        <input type="file" accept=".json,.yaml,.yml,application/json,application/yaml" onchange={pick} />
      </label>
      <label>
        <span>Spec</span>
        <textarea rows="12" bind:value={spec} spellcheck="false" placeholder="openapi: 3.1.0&#10;info: …"></textarea>
      </label>
    </div>
    <footer>
      <button type="button" class="btn" onclick={() => dialog?.close()}>Cancel</button>
      <button type="submit" class="btn primary" disabled={busy}>{busy ? 'Importing…' : 'Import'}</button>
    </footer>
  </form>
</dialog>

<style>
  dialog {
    width: min(40rem, calc(100vw - 2rem));
  }
  dialog header,
  dialog footer {
    padding: 1rem 1.25rem;
  }
  dialog header {
    border-bottom: 1px solid var(--border);
  }
  dialog header h2 {
    margin: 0;
  }
  .body {
    padding: 1.25rem;
    display: grid;
    gap: 0.9rem;
  }
  .body p {
    margin: 0;
    font-size: 0.9rem;
  }
  label {
    display: grid;
    gap: 0.3rem;
  }
  label span {
    font-size: 0.85rem;
    font-weight: 500;
  }
  textarea {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.8rem;
  }
  dialog footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    border-top: 1px solid var(--border);
  }
</style>
