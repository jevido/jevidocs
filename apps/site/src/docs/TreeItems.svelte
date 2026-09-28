<script lang="ts">
  import type { TreeNode } from '../lib/api'
  import Icon, { normalizeIcon } from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import Folder from './TreeFolder.svelte'
  import ApiNav from './api/ApiNav.svelte'
  import { apiNav } from './api/state.svelte'

  let { nodes, depth = 0, onnavigate }: { nodes: TreeNode[]; depth?: number; onnavigate?: () => void } = $props()

  // OpenAPI pages carry their outline as children, but it only exists once
  // the page is open; the chevron says so up front. On the open page it folds
  // the outline, elsewhere it opens the page (and with it the outline).
  let folded = $state<Record<string, boolean>>({})
</script>

{#each nodes as node, i (node.type + ':' + ('slug' in node ? node.slug : node.name) + ':' + i)}
  {#if node.type === 'separator'}
    <p class="separator" class:first={i === 0}>{node.name}</p>
  {:else if node.type === 'page'}
    {@const icon = normalizeIcon(node.icon)}
    {@const current = router.route.slug === node.slug}
    {#if node.kind === 'openapi'}
      {@const open = current && !folded[node.slug]}
      <div class="row" class:current>
        <a
          class="item"
          href={router.href(node.slug)}
          aria-current={current ? 'page' : undefined}
          onclick={() => {
            folded[node.slug] = false
            onnavigate?.()
          }}>
          {#if icon}<Icon name={icon} size={15} />{/if}
          <span>{node.name}</span>
        </a>
        <button
          class="chev"
          type="button"
          aria-label={open ? 'Collapse' : 'Expand'}
          aria-expanded={open}
          onclick={() => {
            if (current) folded[node.slug] = open
            else {
              folded[node.slug] = false
              router.navigate(router.href(node.slug))
            }
          }}>
          <span class:open><Icon name="chevronDown" size={15} /></span>
        </button>
      </div>
      {#if open && apiNav.ref && apiNav.slug === node.slug}
        <ApiNav reference={apiNav.ref} slug={node.slug} {onnavigate} />
      {/if}
    {:else}
      <a
        class="item"
        href={router.href(node.slug)}
        aria-current={current ? 'page' : undefined}
        onclick={() => onnavigate?.()}>
        {#if icon}<Icon name={icon} size={15} />{/if}
        <span>{node.name}</span>
      </a>
    {/if}
  {:else}
    <Folder {node} {depth} {onnavigate} />
  {/if}
{/each}

<style>
  .separator {
    margin: 1.4rem 0 0.4rem;
    padding: 0 0.5rem;
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--fg);
  }
  .separator.first { margin-top: 0.25rem; }
  .item {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.5rem;
    border-radius: var(--radius);
    color: var(--muted-fg);
    font-size: 0.875rem;
    text-decoration: none;
    overflow-wrap: anywhere;
    transition: color 0.12s, background 0.12s;
  }
  .item:hover { background: var(--accent); color: var(--accent-fg); }
  .item[aria-current='page'] {
    background: color-mix(in oklab, var(--brand) 12%, transparent);
    color: var(--brand);
    font-weight: 500;
  }

  /* An OpenAPI page's row, drawn like a folder's: link plus chevron. */
  .row {
    display: flex;
    align-items: center;
    border-radius: var(--radius);
    color: var(--muted-fg);
  }
  .row .item { flex: 1; min-width: 0; }
  .row:hover { background: var(--accent); color: var(--accent-fg); }
  .row:hover .item { background: none; }
  .row.current { background: color-mix(in oklab, var(--brand) 12%, transparent); color: var(--brand); }
  .row.current .item { background: none; }
  .chev {
    display: inline-flex;
    padding: 0.35rem;
    border: 0;
    background: none;
    color: inherit;
    cursor: pointer;
    border-radius: var(--radius);
  }
  .chev span { display: inline-flex; transition: transform 0.15s; transform: rotate(-90deg); }
  .chev span.open { transform: none; }
</style>
