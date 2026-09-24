<script lang="ts">
  import type { TreeNode } from '../lib/api'
  import Icon, { normalizeIcon } from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import Folder from './TreeFolder.svelte'

  let { nodes, depth = 0, onnavigate }: { nodes: TreeNode[]; depth?: number; onnavigate?: () => void } = $props()
</script>

{#each nodes as node, i (node.type + ':' + ('slug' in node ? node.slug : node.name) + ':' + i)}
  {#if node.type === 'separator'}
    <p class="separator" class:first={i === 0}>{node.name}</p>
  {:else if node.type === 'page'}
    {@const icon = normalizeIcon(node.icon)}
    <a
      class="item"
     
      href={router.href(node.slug)}
      aria-current={router.route.slug === node.slug ? 'page' : undefined}
      onclick={() => onnavigate?.()}>
      {#if icon}<Icon name={icon} size={15} />{/if}
      <span>{node.name}</span>
    </a>
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
</style>
