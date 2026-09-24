<script lang="ts">
  import type { TreeNode } from '../lib/api'
  import Icon, { normalizeIcon } from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import { containsSlug } from '../lib/tree'
  import TreeItems from './TreeItems.svelte'

  type FolderNode = Extract<TreeNode, { type: 'folder' }>
  let { node, depth, onnavigate }: { node: FolderNode; depth: number; onnavigate?: () => void } = $props()

  const active = $derived(containsSlug(node, router.route.slug))
  let open = $state(false)
  let touched = false

  // Opens by default when configured, and whenever navigation lands inside.
  $effect.pre(() => {
    if (active || (!touched && node.defaultOpen)) open = true
  })

  const icon = $derived(normalizeIcon(node.icon))
  const toggle = () => {
    touched = true
    open = !open
  }
</script>

<div class="folder">
  <div class="row">
    {#if node.index}
      <a
        class="label"
        href={router.href(node.index.slug)}
        aria-current={router.route.slug === node.index.slug ? 'page' : undefined}
        onclick={() => {
          open = true
          onnavigate?.()
        }}>
        {#if icon}<Icon name={icon} size={15} />{/if}
        <span>{node.name}</span>
      </a>
    {:else}
      <button class="label" type="button" onclick={toggle} aria-expanded={open}>
        {#if icon}<Icon name={icon} size={15} />{/if}
        <span>{node.name}</span>
      </button>
    {/if}
    <button class="chev" type="button" aria-label={open ? 'Collapse' : 'Expand'} aria-expanded={open} onclick={toggle}>
      <span class:open><Icon name="chevronDown" size={15} /></span>
    </button>
  </div>
  {#if open}
    <div class="children">
      <TreeItems nodes={node.children} depth={depth + 1} {onnavigate} />
    </div>
  {/if}
</div>

<style>
  .row {
    display: flex;
    align-items: center;
    border-radius: var(--radius);
    color: var(--muted-fg);
  }
  .row:hover { background: var(--accent); color: var(--accent-fg); }
  .label {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    padding: 0.4rem 0.5rem;
    border: 0;
    background: none;
    color: inherit;
    font-size: 0.875rem;
    text-align: left;
    text-decoration: none;
    cursor: pointer;
    border-radius: var(--radius);
  }
  .label[aria-current='page'] { color: var(--brand); font-weight: 500; }
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
  .children {
    position: relative;
    margin-left: 0.9rem;
    padding-left: 0.35rem;
  }
  .children::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0.25rem;
    bottom: 0.25rem;
    width: 1px;
    background: var(--border);
  }
</style>
