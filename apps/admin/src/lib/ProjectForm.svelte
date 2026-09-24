<script lang="ts">
  import type { ProjectInput } from './types'
  import { slugify } from './format'
  import { untrack } from 'svelte'

  let {
    value = $bindable(),
    lockSlug = false,
  }: { value: ProjectInput; lockSlug?: boolean } = $props()

  let slugTouched = $state(untrack(() => lockSlug))

  function onName() {
    if (!slugTouched) value.slug = slugify(value.name)
  }
</script>

<div class="grid">
  <label class="field">
    Name
    <input type="text" required bind:value={value.name} oninput={onName} />
  </label>
  <label class="field">
    Slug
    <input
      type="text"
      required
      pattern="[a-z0-9][a-z0-9-]*"
      bind:value={value.slug}
      oninput={() => (slugTouched = true)}
      disabled={lockSlug}
    />
    <span class="hint">Lowercase letters, digits and dashes. Used in URLs.</span>
  </label>
  <label class="field wide">
    Description
    <input type="text" bind:value={value.description} />
  </label>
  <label class="field wide">
    GitHub URL
    <input type="url" placeholder="https://github.com/you/repo" bind:value={value.github_url} />
  </label>
  <label class="field wide">
    Edit URL
    <input type="text" placeholder="https://github.com/you/repo/blob/main/docs/{'{path}'}" bind:value={value.edit_url} />
    <span class="hint">"Edit this page" link. <code>{'{path}'}</code> becomes the page's file, e.g. <code>guides/index.md</code>.</span>
  </label>
  <label class="field wide">
    Banner
    <input type="text" placeholder="Announcement shown above every page" bind:value={value.banner} />
  </label>
  <div class="field wide">
    <span>Navbar links</span>
    {#each value.links as link, i (i)}
      <div class="row">
        <input type="text" placeholder="Text" bind:value={link.text} />
        <input type="text" placeholder="https://…" bind:value={link.url} />
        <button type="button" class="btn sm ghost danger" onclick={() => value.links.splice(i, 1)} aria-label="Remove link">✕</button>
      </div>
    {/each}
    <div>
      <button type="button" class="btn sm" onclick={() => value.links.push({ text: '', url: '' })}>+ Add link</button>
    </div>
  </div>
  <label class="check wide">
    <input type="checkbox" bind:checked={value.public} />
    Public: listed and readable without signing in
  </label>
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
  .wide {
    grid-column: 1 / -1;
  }
  .field .row {
    margin-bottom: 0.4rem;
  }
  @media (max-width: 640px) {
    .grid {
      grid-template-columns: 1fr;
    }
  }
</style>
