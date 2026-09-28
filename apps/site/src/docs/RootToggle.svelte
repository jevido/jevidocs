<script lang="ts">
  import type { TreeNode } from '../lib/api'
  import Icon, { normalizeIcon } from '../lib/Icon.svelte'
  import { router } from '../lib/router.svelte'
  import { apiPageOf, flatten } from '../lib/tree'

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
      icon: normalizeIcon(r.icon) ?? (apiPageOf(r) ? 'code' : 'folder'),
      // An API root opens on its reference, not on its overview.
      href: router.href(apiPageOf(r)?.slug ?? r.index?.slug ?? firstSlug(r.children) ?? ''),
      active: active === r,
    })),
  ])
  const current = $derived(options.find((o) => o.active) ?? options[0])

  // The menu is a windowed list: five rows tall, rendering only the rows in
  // view (plus a few either side), so projects with many roots stay cheap.
  const ROW = 44 // px, matches `.menu a` height
  const VISIBLE = 5
  const OVERSCAN = 2
  let scrollTop = $state(0)
  const start = $derived(Math.max(0, Math.floor(scrollTop / ROW) - OVERSCAN))
  const end = $derived(Math.min(options.length, Math.ceil(scrollTop / ROW) + VISIBLE + OVERSCAN))
  const shown = $derived(options.slice(start, end))

  // Opens scrolled to the active root.
  function reveal(node: HTMLElement) {
    const i = options.findIndex((o) => o.active)
    if (i >= VISIBLE) node.scrollTop = (i - Math.floor(VISIBLE / 2)) * ROW
    scrollTop = node.scrollTop
  }

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
    <div
      class="menu"
      role="menu"
      style:--row="{ROW}px"
      style:max-height="{VISIBLE * ROW}px"
      onscroll={(e) => (scrollTop = e.currentTarget.scrollTop)}
      {@attach reveal}>
      <div class="rows" style:height="{options.length * ROW}px">
        {#each shown as o, i (o.href + o.name)}
          <a
            role="menuitem"
            href={o.href}
            class:active={o.active}
            style:top="{(start + i) * ROW}px"
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
    overflow-y: auto;
    overscroll-behavior: contain;
    box-sizing: content-box;
  }
  .rows { position: relative; }
  .menu a {
    position: absolute;
    left: 0;
    right: 0;
    height: var(--row);
    box-sizing: border-box;
  }
  .menu a:hover { background: var(--accent); }
  .menu a.active { background: var(--accent); }
</style>
