<script lang="ts">
  import type { TreeNode } from '../lib/api'
  import Icon, { normalizeIcon } from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import { flatten } from '../lib/tree'

  type Folder = Extract<TreeNode, { type: 'folder' }>
  type Option = { name: string; description: string; icon?: string; href: string; active: boolean }

  let {
    projectName,
    roots,
    active,
    rest,
    onnavigate,
  }: { projectName: string; roots: Folder[]; active: Folder | null; rest: TreeNode[]; onnavigate?: () => void } =
    $props()

  let open = $state(false)

  function firstSlug(nodes: TreeNode[]): string | undefined {
    return flatten(nodes)[0]?.slug
  }

  const options = $derived<Option[]>([
    {
      name: projectName,
      description: 'Main documentation',
      icon: 'book',
      href: router.href(firstSlug(rest) ?? ''),
      active: active === null,
    },
    ...roots.map((r) => ({
      name: r.name,
      description: r.description ?? '',
      icon: normalizeIcon(r.icon) ?? 'folder',
      href: router.href(r.index?.slug ?? firstSlug(r.children) ?? ''),
      active: active === r,
    })),
  ])
  const current = $derived(options.find((o) => o.active) ?? options[0])

  function close(e: MouseEvent) {
    if (!(e.target as HTMLElement).closest('.root-toggle')) open = false
  }
</script>

<svelte:window onclick={close} onkeydown={(e) => e.key === 'Escape' && (open = false)} />

<div class="root-toggle">
  <button type="button" class="trigger" aria-expanded={open} onclick={() => (open = !open)}>
    <span class="icon"><Icon name={current.icon ?? 'book'} size={16} /></span>
    <span class="text">
      <span class="name">{current.name}</span>
      {#if current.description}<span class="desc">{current.description}</span>{/if}
    </span>
    <Icon name="chevronDown" size={15} />
  </button>
  {#if open}
    <div class="menu" role="menu">
      {#each options as o (o.href + o.name)}
        <a
          role="menuitem"
          href={o.href}
          class:active={o.active}
          onclick={() => {
            open = false
            onnavigate?.()
          }}>
          <span class="icon"><Icon name={o.icon ?? 'folder'} size={16} /></span>
          <span class="text">
            <span class="name">{o.name}</span>
            {#if o.description}<span class="desc">{o.description}</span>{/if}
          </span>
          {#if o.active}<Icon name="check" size={15} />{/if}
        </a>
      {/each}
    </div>
  {/if}
</div>

<style>
  .root-toggle { position: relative; margin: 0 0 0.75rem; }
  .trigger,
  .menu a {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    width: 100%;
    padding: 0.45rem 0.55rem;
    border-radius: var(--radius);
    text-align: left;
    color: var(--fg);
    text-decoration: none;
    font: inherit;
  }
  .trigger {
    border: 1px solid var(--border);
    background: var(--card);
    cursor: pointer;
  }
  .trigger:hover { background: var(--accent); }
  .icon {
    display: grid;
    place-items: center;
    width: 1.9rem;
    height: 1.9rem;
    flex: none;
    border-radius: 0.4rem;
    border: 1px solid var(--border);
    background: var(--bg);
    color: var(--brand);
  }
  .text { display: flex; flex-direction: column; min-width: 0; flex: 1; }
  .name { font-size: 0.875rem; font-weight: 500; }
  .desc {
    font-size: 0.75rem;
    color: var(--muted-fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .menu {
    position: absolute;
    z-index: 30;
    top: calc(100% + 0.35rem);
    left: 0;
    right: 0;
    padding: 0.3rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--popover);
    box-shadow: var(--shadow);
  }
  .menu a:hover { background: var(--accent); }
  .menu a.active { background: var(--accent); }
</style>
