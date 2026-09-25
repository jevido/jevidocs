<script lang="ts">
  import type { ApiReference } from '../../lib/openapi'
  import { router } from '../../lib/router.svelte'
  import MethodBadge from './MethodBadge.svelte'
  import { apiNav } from './state.svelte'

  // The open OpenAPI page's outline under its sidebar entry: tags with their
  // operations, then models, following the scroll position.
  let { reference, slug, onnavigate }: { reference: ApiReference; slug: string; onnavigate?: () => void } = $props()

  const href = (id: string) => `${router.href(slug)}#${id}`
  const activeTag = $derived(
    reference.tags.find((t) => t.id === apiNav.active || t.operations.some((o) => o.id === apiNav.active))?.id ??
      (apiNav.active.startsWith('model/') || apiNav.active === 'models' ? 'models' : ''),
  )
  let closed = $state<Record<string, boolean>>({})
  const isOpen = (id: string) => (id in closed ? !closed[id] : id === activeTag || reference.tags.length <= 3)
</script>

<nav class="api-nav" aria-label="{reference.title} endpoints">
  {#each reference.tags as tag (tag.id)}
    {@const open = isOpen(tag.id)}
    <div class="group">
      <button class="tag" type="button" aria-expanded={open} onclick={() => (closed[tag.id] = open)}>
        <span class="chev" class:open aria-hidden="true">›</span>
        <span>{tag.name}</span>
      </button>
      {#if open}
        {#each tag.operations as op (op.id)}
          <a class="op" href={href(op.id)} aria-current={apiNav.active === op.id ? 'location' : undefined} onclick={() => onnavigate?.()}>
            <span class="name" class:deprecated={op.deprecated}>{op.summary}</span>
            <MethodBadge method={op.method} small />
          </a>
        {/each}
      {/if}
    </div>
  {/each}
  {#if reference.models.length}
    {@const open = isOpen('models')}
    <div class="group">
      <button class="tag" type="button" aria-expanded={open} onclick={() => (closed.models = open)}>
        <span class="chev" class:open aria-hidden="true">›</span>
        <span>Models</span>
      </button>
      {#if open}
        {#each reference.models as m (m.id)}
          <a class="op" href={href(m.id)} aria-current={apiNav.active === m.id ? 'location' : undefined} onclick={() => onnavigate?.()}>
            <span class="name mono">{m.name}</span>
          </a>
        {/each}
      {/if}
    </div>
  {/if}
</nav>

<style>
  .api-nav { margin: 0.15rem 0 0.4rem 0.6rem; padding-left: 0.5rem; border-left: 1px solid var(--border); }
  .tag {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    width: 100%;
    padding: 0.35rem 0.4rem;
    border: 0;
    border-radius: var(--radius);
    background: none;
    font-size: 0.82rem;
    font-weight: 500;
    color: var(--fg);
    text-align: left;
    cursor: pointer;
  }
  .tag:hover { background: var(--accent); }
  .chev { display: inline-block; width: 0.7rem; color: var(--muted-fg); transition: transform 0.12s; }
  .chev.open { transform: rotate(90deg); }
  .op {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    padding: 0.3rem 0.4rem 0.3rem 1.45rem;
    border-radius: var(--radius);
    font-size: 0.8rem;
    color: var(--muted-fg);
    text-decoration: none;
  }
  .op:hover { background: var(--accent); color: var(--accent-fg); }
  .op[aria-current] { color: var(--fg); background: var(--muted); font-weight: 500; }
  .name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .name.deprecated { text-decoration: line-through; }
  .mono { font-family: var(--font-mono); font-size: 0.76rem; }
</style>
