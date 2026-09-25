<script lang="ts">
  import { router } from '../../lib/router.svelte'
  import { resolve, typeLabel } from '../../lib/openapi'
  import SchemaView from './SchemaView.svelte'
  import { getApi } from './state.svelte'

  let { model }: { model: { id: string; name: string } } = $props()
  const api = getApi()
  const s = $derived(resolve({ ref: model.name }, api.ref.schemas))
  let open = $state(false)

  // Opens when the address points at it, e.g. from search or the sidebar.
  $effect.pre(() => {
    if (decodeURIComponent(router.hash.slice(1)) === model.id) open = true
  })
</script>

<details class="model" id={model.id} data-model bind:open>
  <summary>
    <span class="name">{model.name}</span>
    <span class="type">{typeLabel(s, api.ref.schemas)}</span>
  </summary>
  {#if open}
    <div class="body">
      {#if s?.description}<div class="prose desc">{@html s.description}</div>{/if}
      <SchemaView schema={{ ref: model.name }} />
    </div>
  {/if}
</details>

<style>
  .model { border: 1px solid var(--border); border-radius: 0.6rem; max-width: calc(50% - 1.5rem); }
  summary {
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
    padding: 0.65rem 0.9rem;
    cursor: pointer;
    list-style: none;
  }
  summary::-webkit-details-marker { display: none; }
  summary::before { content: '+'; font-family: var(--font-mono); color: var(--muted-fg); width: 0.7rem; }
  .model[open] summary::before { content: '−'; }
  summary:hover { background: var(--muted); border-radius: 0.6rem; }
  .name { font-family: var(--font-mono); font-size: 0.85rem; font-weight: 600; }
  .type { font-family: var(--font-mono); font-size: 0.75rem; color: var(--muted-fg); }
  .body { padding: 0 0.9rem 0.5rem; border-top: 1px solid var(--border-soft); }
  .desc { font-size: 0.85rem; }
  @media (max-width: 1100px) {
    .model { max-width: none; }
  }
</style>
