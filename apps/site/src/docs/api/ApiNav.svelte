<script lang="ts">
  import type { ApiReference } from '../../lib/openapi'
  import Icon from '../../lib/Icon.svelte'
  import { router } from '../../lib/router.svelte'
  import MethodBadge from './MethodBadge.svelte'
  import { apiNav } from './state.svelte'

  // The open OpenAPI page's outline under its sidebar entry: tags with their
  // operations, then models, following the scroll position. `flat` is the
  // outline as a root's own sidebar, drawn like the page tree's folders.
  let {
    reference,
    slug,
    flat = false,
    onnavigate,
  }: { reference: ApiReference; slug: string; flat?: boolean; onnavigate?: () => void } = $props()

  const href = (id: string) => `${router.href(slug)}#${id}`
  const activeTag = $derived(
    reference.tags.find((t) => t.id === apiNav.active || t.operations.some((o) => o.id === apiNav.active))?.id ??
      (apiNav.active.startsWith('model/') || apiNav.active === 'models' ? 'models' : ''),
  )
  // A small API stays unfolded; a large one opens only where the reader is.
  const small = $derived(reference.tags.length <= 3 || reference.tags.reduce((n, t) => n + t.operations.length, 0) <= 20)
  let closed = $state<Record<string, boolean>>({})
  const isOpen = (id: string) => (id in closed ? !closed[id] : id === activeTag || small)
</script>

<nav class={['api-nav', { flat }]} aria-label="{reference.title} endpoints">
  {#each reference.tags as tag (tag.id)}
    {@const open = isOpen(tag.id)}
    <div class="group">
      <button class="tag" type="button" aria-expanded={open} onclick={() => (closed[tag.id] = open)}>
        {@render chev(open)}
        <span class="label">{tag.name}</span>
      </button>
      {#if open}
        <div class="ops">
          {#each tag.operations as op (op.id)}
            <a class="op" href={href(op.id)} aria-current={apiNav.active === op.id ? 'location' : undefined} onclick={() => onnavigate?.()}>
              <span class="name" class:deprecated={op.deprecated}>{op.summary}</span>
              <MethodBadge method={op.method} small />
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/each}
  {#if reference.models.length}
    {@const open = isOpen('models')}
    <div class="group">
      <button class="tag" type="button" aria-expanded={open} onclick={() => (closed.models = open)}>
        {@render chev(open)}
        <span class="label">Models</span>
      </button>
      {#if open}
        <div class="ops">
          {#each reference.models as m (m.id)}
            <a class="op" href={href(m.id)} aria-current={apiNav.active === m.id ? 'location' : undefined} onclick={() => onnavigate?.()}>
              <span class="name mono">{m.name}</span>
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</nav>

{#snippet chev(open: boolean)}
  {#if flat}
    <span class="fold" class:open aria-hidden="true"><Icon name="chevronDown" size={15} /></span>
  {:else}
    <span class="chev" class:open aria-hidden="true">›</span>
  {/if}
{/snippet}

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

  /* As a root's sidebar: tags look like folders of the page tree. */
  .flat { margin: 0; padding: 0; border: 0; }
  .flat .tag {
    padding: 0.4rem 0.5rem;
    font-size: 0.875rem;
    font-weight: 400;
    color: var(--muted-fg);
  }
  .flat .tag:hover { color: var(--accent-fg); }
  .flat .label { flex: 1; }
  .fold { display: inline-flex; order: 1; transition: transform 0.15s; transform: rotate(-90deg); }
  .fold.open { transform: none; }
  .flat .ops { position: relative; margin-left: 0.9rem; padding-left: 0.35rem; }
  .flat .ops::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0.25rem;
    bottom: 0.25rem;
    width: 1px;
    background: var(--border);
  }
  .flat .op { padding: 0.4rem 0.5rem; font-size: 0.875rem; }
  .flat .op[aria-current] {
    background: color-mix(in oklab, var(--brand) 12%, transparent);
    color: var(--brand);
  }
</style>
