<script lang="ts">
  import type { Snippet } from 'svelte'
  import Icon from '../../lib/Icon.svelte'
  import { highlight } from '../../lib/openapi'

  let {
    code,
    lang,
    head,
    foot,
    empty = 'No body',
  }: { code: string; lang: string; head?: Snippet; foot?: Snippet; empty?: string } = $props()

  let copied = $state(false)
  let timer: ReturnType<typeof setTimeout> | undefined

  async function copy() {
    try {
      await navigator.clipboard.writeText(code)
      copied = true
      clearTimeout(timer)
      timer = setTimeout(() => (copied = false), 1500)
    } catch {
      // clipboard blocked: nothing to do
    }
  }
</script>

<div class="panel">
  {#if head}
    <div class="head">{@render head()}</div>
  {/if}
  <div class="body">
    {#if code}
      <pre class="chroma"><code>{@html highlight(code, lang)}</code></pre>
      <button class="copy" type="button" aria-label={copied ? 'Copied' : 'Copy code'} data-copied={copied || undefined} onclick={copy}>
        <Icon name={copied ? 'check' : 'copy'} size={14} />
      </button>
    {:else}
      <p class="empty">{empty}</p>
    {/if}
  </div>
  {#if foot}
    <div class="foot">{@render foot()}</div>
  {/if}
</div>

<style>
  .panel {
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    background: var(--card);
    overflow: hidden;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 2.6rem;
    padding: 0.35rem 0.5rem 0.35rem 0.85rem;
    border-bottom: 1px solid var(--border);
    background: var(--muted);
    font-size: 0.8rem;
  }
  .body { position: relative; }
  pre {
    margin: 0;
    max-height: 26rem;
    padding: 0.85rem 2.75rem 0.85rem 1rem;
    overflow: auto;
    font-family: var(--font-mono);
    font-size: 0.78rem;
    line-height: 1.65;
    scrollbar-width: thin;
  }
  code { font-family: inherit; }
  .copy {
    position: absolute;
    top: 0.4rem;
    right: 0.4rem;
    display: inline-grid;
    place-items: center;
    width: 1.8rem;
    height: 1.8rem;
    border: 0;
    border-radius: 6px;
    background: var(--card);
    color: var(--muted-fg);
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.12s, background 0.12s;
  }
  .body:hover .copy,
  .copy:focus-visible,
  .copy[data-copied] { opacity: 1; }
  .copy:hover { background: var(--accent); color: var(--fg); }
  .copy[data-copied] { color: var(--success); }
  .empty {
    margin: 0;
    padding: 1.25rem 1rem;
    font-size: 0.8rem;
    color: var(--muted-fg);
    text-align: center;
  }
  .foot {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.4rem 0.5rem;
    border-top: 1px solid var(--border);
  }
  @media (hover: none) {
    .copy { opacity: 1; }
  }
</style>
