<script lang="ts">
  import { hasChildren, resolve, typeLabel, type ApiSchema } from '../../lib/openapi'
  import SchemaView from './SchemaView.svelte'
  import { getApi } from './state.svelte'

  // One attribute row, Scalar-style: name, type and flags, description,
  // allowed values, and a toggle for nested attributes.
  let {
    name,
    schema,
    required = false,
    depth = 0,
    description,
    deprecated = false,
    example,
  }: {
    name: string
    schema?: ApiSchema
    required?: boolean
    depth?: number
    description?: string
    deprecated?: boolean
    example?: unknown
  } = $props()

  const api = getApi()
  const s = $derived(resolve(schema, api.ref.schemas))
  const type = $derived(typeLabel(schema, api.ref.schemas))
  const desc = $derived(description || schema?.description || s?.description || '')
  const nested = $derived(hasChildren(schema, api.ref.schemas))
  let open = $state(false)

  const show = (v: unknown) => (typeof v === 'string' ? v : JSON.stringify(v))
</script>

<li class="prop">
  <div class="line">
    <code class="name" class:deprecated={deprecated || s?.deprecated}>{name}</code>
    <span class="type">{type}</span>
    {#if s?.nullable}<span class="flag">nullable</span>{/if}
    {#if s?.read_only}<span class="flag">read-only</span>{/if}
    {#if s?.write_only}<span class="flag">write-only</span>{/if}
    {#if deprecated || s?.deprecated}<span class="flag warn">deprecated</span>{/if}
    {#each s?.constraints ?? [] as c (c)}<span class="flag">{c}</span>{/each}
    {#if required}<span class="required">required</span>{/if}
  </div>
  {#if desc}<div class="desc prose">{@html desc}</div>{/if}
  {#if s?.enum?.length}
    <div class="values">
      <span class="label">{s.enum.length === 1 ? 'Value' : 'Values'}</span>
      {#each s.enum as v, i (i)}<code>{show(v)}</code>{/each}
    </div>
  {/if}
  {#if s?.default !== undefined}
    <div class="values"><span class="label">Default</span><code>{show(s.default)}</code></div>
  {/if}
  {#if (example ?? s?.example) !== undefined && !s?.enum?.length && show(example ?? s?.example) !== show(s?.default)}
    <div class="values"><span class="label">Example</span><code>{show(example ?? s?.example)}</code></div>
  {/if}
  {#if nested}
    <div class="children" class:open>
      <button type="button" class="toggle" aria-expanded={open} onclick={() => (open = !open)}>
        <span class="plus" aria-hidden="true">{open ? '−' : '+'}</span>
        {open ? 'Hide' : 'Show'} child attributes
      </button>
      {#if open}
        <div class="nested"><SchemaView schema={schema!} depth={depth + 1} /></div>
      {/if}
    </div>
  {/if}
</li>

<style>
  .prop {
    padding: 0.75rem 0;
    border-top: 1px solid var(--border-soft);
    font-size: 0.875rem;
  }
  .prop:first-child { border-top: 0; }
  .line { display: flex; align-items: baseline; flex-wrap: wrap; gap: 0.25rem 0.5rem; }
  .name {
    font-family: var(--font-mono);
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .name.deprecated { text-decoration: line-through; color: var(--muted-fg); }
  .type { font-family: var(--font-mono); font-size: 0.75rem; color: var(--muted-fg); overflow-wrap: anywhere; }
  .flag {
    font-family: var(--font-mono);
    font-size: 0.7rem;
    color: var(--muted-fg);
    padding: 0 0.35rem;
    border-radius: 4px;
    background: var(--muted);
  }
  .flag.warn { color: var(--warn); }
  .required { font-size: 0.72rem; font-weight: 500; color: var(--warn); }
  .desc { margin-top: 0.3rem; font-size: 0.85rem; color: var(--muted-fg); }
  .desc :global(p) { margin: 0.2rem 0; }
  .values { display: flex; flex-wrap: wrap; align-items: center; gap: 0.3rem; margin-top: 0.4rem; font-size: 0.75rem; }
  .values .label { color: var(--muted-fg); margin-right: 0.15rem; }
  .values code {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    padding: 0.05rem 0.35rem;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--card);
    overflow-wrap: anywhere;
  }
  .children {
    margin-top: 0.6rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    width: fit-content;
  }
  .children.open { border-radius: 0.6rem; width: auto; }
  .toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.25rem 0.75rem;
    border: 0;
    background: none;
    font-size: 0.78rem;
    color: var(--muted-fg);
    cursor: pointer;
  }
  .toggle:hover { color: var(--fg); }
  .children.open .toggle { width: 100%; border-bottom: 1px solid var(--border); }
  .plus { font-family: var(--font-mono); width: 0.7rem; }
  .nested { padding: 0 0.85rem; }
</style>
