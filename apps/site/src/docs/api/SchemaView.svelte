<script lang="ts">
  import { resolve, typeLabel, type ApiSchema } from '../../lib/openapi'
  import SchemaProp from './SchemaProp.svelte'
  import Self from './SchemaView.svelte'
  import { getApi } from './state.svelte'

  // The attributes of a schema: its properties, an array's item attributes,
  // or the variants of a oneOf/anyOf with a switcher.
  let { schema, depth = 0 }: { schema: ApiSchema; depth?: number } = $props()

  const api = getApi()
  const s = $derived(resolve(schema, api.ref.schemas))
  let variant = $state(0)
  const variants = $derived(s?.variants ?? [])
  const labels = $derived(
    variants.map((v, i) => resolve(v, api.ref.schemas)?.title || v.ref || (typeLabel(v, api.ref.schemas) !== 'object' ? typeLabel(v, api.ref.schemas) : `Option ${i + 1}`)),
  )
</script>

{#if !s}
  <p class="note">Unknown schema</p>
{:else if s.truncated || depth > 14}
  <p class="note">Nested too deeply to show.</p>
{:else}
  {#if variants.length > 1}
    <div class="variants">
      <span class="kind">{s.composition === 'anyOf' ? 'Any of' : s.composition === 'allOf' ? 'All of' : 'One of'}</span>
      <div class="tabs" role="tablist">
        {#each labels as label, i (i)}
          <button type="button" role="tab" aria-selected={variant === i} onclick={() => (variant = i)}>{label}</button>
        {/each}
      </div>
    </div>
    {#if variants[variant]}
      {@const v = resolve(variants[variant], api.ref.schemas)}
      {#if v?.description}<div class="vdesc prose">{@html v.description}</div>{/if}
      <Self schema={variants[variant]} depth={depth + 1} />
    {/if}
  {:else if variants.length === 1}
    <Self schema={variants[0]} depth={depth + 1} />
  {/if}
  {#if s.properties?.length}
    <ul class="props">
      {#each s.properties as p (p.name)}
        <SchemaProp name={p.name} required={p.required} schema={p.schema} {depth} />
      {/each}
    </ul>
  {/if}
  {#if s.additional_properties}
    <ul class="props">
      <SchemaProp name="[key: string]" schema={s.additional_properties} {depth} />
    </ul>
  {/if}
  {#if s.items && !s.properties?.length && !variants.length}
    <Self schema={s.items} depth={depth + 1} />
  {/if}
{/if}

<style>
  .props { list-style: none; margin: 0; padding: 0; }
  .note { margin: 0.5rem 0; font-size: 0.8rem; color: var(--muted-fg); }
  .variants {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0.6rem 0;
    border-bottom: 1px solid var(--border-soft);
  }
  .kind { font-size: 0.75rem; font-weight: 500; color: var(--muted-fg); }
  .tabs { display: flex; flex-wrap: wrap; gap: 0.25rem; }
  .tabs button {
    padding: 0.15rem 0.55rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: transparent;
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--muted-fg);
    cursor: pointer;
  }
  .tabs button[aria-selected='true'] { background: var(--accent); color: var(--fg); border-color: transparent; }
  .vdesc { margin-top: 0.5rem; font-size: 0.85rem; }
</style>
