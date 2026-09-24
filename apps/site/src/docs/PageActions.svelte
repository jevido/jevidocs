<script lang="ts">
  import Icon from '../lib/Icon.svelte'
  import { copyText } from '../lib/enhance'

  let { markdown, markdownUrl }: { markdown: string; markdownUrl: string } = $props()

  let copied = $state(false)
  let menuOpen = $state(false)
  let root = $state<HTMLElement>()

  const prompt = $derived(`Read ${markdownUrl}, I want to ask questions about it.`)
  const targets = $derived([
    { name: 'Open in ChatGPT', url: `https://chatgpt.com/?hints=search&q=${encodeURIComponent(prompt)}` },
    { name: 'Open in Claude', url: `https://claude.ai/new?q=${encodeURIComponent(prompt)}` },
    { name: 'Open in Scira AI', url: `https://scira.ai/?q=${encodeURIComponent(prompt)}` },
    { name: 'View as Markdown', url: markdownUrl },
  ])

  async function copy() {
    if (await copyText(markdown)) {
      copied = true
      setTimeout(() => (copied = false), 1500)
    }
  }
</script>

<svelte:window
  onclick={(e) => {
    if (menuOpen && root && !root.contains(e.target as Node)) menuOpen = false
  }}
  onkeydown={(e) => {
    if (e.key === 'Escape') menuOpen = false
  }} />

<div class="actions" bind:this={root}>
  <button type="button" class="pill" onclick={copy}>
    <Icon name={copied ? 'check' : 'copy'} size={14} />
    {copied ? 'Copied' : 'Copy Markdown'}
  </button>
  <div class="menu-wrap">
    <button type="button" class="pill" aria-haspopup="menu" aria-expanded={menuOpen} onclick={() => (menuOpen = !menuOpen)}>
      Open
      <Icon name="chevronDown" size={14} />
    </button>
    {#if menuOpen}
      <div class="menu" role="menu">
        {#each targets as t (t.name)}
          <a role="menuitem" href={t.url} target="_blank" rel="noreferrer noopener" onclick={() => (menuOpen = false)}>
            {t.name}
            <Icon name="external" size={13} />
          </a>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
  .pill {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    height: 2rem;
    padding: 0 0.75rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
    color: var(--muted-fg);
    font-size: 0.8rem;
    font-weight: 500;
    cursor: pointer;
  }
  .pill:hover { background: var(--accent); color: var(--accent-fg); }
  .menu-wrap { position: relative; }
  .menu {
    position: absolute;
    z-index: 30;
    top: calc(100% + 0.35rem);
    left: 0;
    min-width: 12rem;
    padding: 0.3rem;
    background: var(--popover);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }
  .menu a {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.45rem 0.6rem;
    border-radius: calc(var(--radius) - 2px);
    font-size: 0.85rem;
    color: var(--fg);
    text-decoration: none;
  }
  .menu a:hover { background: var(--accent); }
</style>
