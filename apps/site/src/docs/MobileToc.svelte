<script lang="ts">
  import type { TocItem } from '../lib/api'
  import Icon from '../lib/Icon.svelte'

  // "On this page" for narrow screens, where the TOC column is hidden.
  let { items }: { items: TocItem[] } = $props()
  let open = $state(false)
</script>

{#if items.length > 0}
  <div class="mtoc">
    <button type="button" aria-expanded={open} onclick={() => (open = !open)}>
      <Icon name="text" size={15} />
      <span>On this page</span>
      <span class="chev" class:open><Icon name="chevronDown" size={15} /></span>
    </button>
    {#if open}
      <ul>
        {#each items as item (item.url)}
          <li style:padding-left="{(item.depth - 2) * 0.85}rem">
            <a href={item.url} onclick={() => (open = false)}>{item.title}</a>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .mtoc {
    display: none;
    margin: 0 0 1.25rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
  }
  button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    padding: 0.6rem 0.8rem;
    border: 0;
    background: none;
    color: var(--muted-fg);
    font: inherit;
    font-size: 0.875rem;
    cursor: pointer;
  }
  .chev { margin-left: auto; display: flex; transition: transform 0.15s; }
  .chev.open { transform: rotate(180deg); }
  ul { list-style: none; margin: 0; padding: 0 0.8rem 0.7rem; }
  li a {
    display: block;
    padding: 0.25rem 0;
    font-size: 0.85rem;
    color: var(--muted-fg);
    text-decoration: none;
  }
  li a:hover { color: var(--fg); }
  @media (max-width: 1200px) {
    .mtoc { display: block; }
  }
  @media print {
    .mtoc { display: none; }
  }
</style>
