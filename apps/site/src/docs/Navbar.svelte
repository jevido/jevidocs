<script lang="ts">
  import type { Project } from '../lib/api'
  import Icon from '../lib/Icon.svelte'
  import Logo from '../lib/Logo.svelte'
  import ThemeToggle from '../lib/ThemeToggle.svelte'
  import { router } from '../lib/router.svelte'

  let {
    project,
    onsearch,
    onmenu,
  }: { project: Project | null; onsearch: () => void; onmenu: () => void } = $props()

  const isMac = typeof navigator !== 'undefined' && /mac|iphone|ipad/i.test(navigator.platform || navigator.userAgent)
  const links = $derived(project?.links ?? [])
  const versions = $derived(project?.versions ?? [])

  // Switch version and stay on the same page; the reader shows its
  // not-found state if that page does not exist in the other version.
  function switchVersion(e: Event) {
    const slug = (e.currentTarget as HTMLSelectElement).value
    const base = slug === 'jevidocs' ? '/docs' : `/p/${slug}`
    router.navigate(router.href(router.route.slug, base))
  }
</script>

<header class="nav">
  <div class="inner">
    <button class="icon-btn menu" type="button" aria-label="Open navigation" onclick={onmenu}>
      <Icon name="menu" size={18} />
    </button>
    <a class="brand" href={router.href('')}>
      {#if project?.logo_url}
        <img class="project-logo" src={project.logo_url} alt="" width="22" height="22" />
      {:else}
        <Logo />
      {/if}
      <span>{project?.name ?? 'Docs'}</span>
    </a>
    {#if versions.length > 1 && project}
      <select class="version" aria-label="Documentation version" value={project.slug} onchange={switchVersion}>
        {#each versions as v (v.slug)}
          <option value={v.slug}>{v.label}</option>
        {/each}
      </select>
    {/if}
    <a class="home" href="/">jevidocs</a>

    <button class="search" type="button" onclick={onsearch}>
      <Icon name="search" size={15} />
      <span class="label">Search</span>
      <span class="keys"><kbd>{isMac ? '⌘' : 'Ctrl'}</kbd><kbd>K</kbd></span>
    </button>

    <nav class="links" aria-label="Project links">
      {#each links as l (l.url)}
        <a href={l.url}>{l.text}</a>
      {/each}
    </nav>
    <div class="end">
      <button class="icon-btn search-sm" type="button" aria-label="Search" onclick={onsearch}>
        <Icon name="search" size={18} />
      </button>
      {#if project?.github_url}
        <a class="icon-btn" href={project.github_url} target="_blank" rel="noreferrer noopener" aria-label="GitHub">
          <Icon name="github" size={18} />
        </a>
      {/if}
      <ThemeToggle />
    </div>
  </div>
</header>

<style>
  .nav {
    position: sticky;
    top: 0;
    z-index: 40;
    height: var(--nav-h);
    border-bottom: 1px solid var(--border);
    background: color-mix(in oklab, var(--bg) 80%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
  .inner {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    height: 100%;
    max-width: 1400px;
    margin: 0 auto;
    padding: 0 1rem;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    font-weight: 600;
    font-size: 0.95rem;
    text-decoration: none;
    white-space: nowrap;
  }
  .home {
    font-size: 0.8rem;
    color: var(--muted-fg);
    text-decoration: none;
    padding: 0.15rem 0.5rem;
    border: 1px solid var(--border);
    border-radius: 999px;
  }
  .home:hover { color: var(--fg); }
  .search {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    width: 16rem;
    height: 2.2rem;
    margin-left: 1rem;
    padding: 0 0.5rem 0 0.75rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--card);
    color: var(--muted-fg);
    font-size: 0.85rem;
    cursor: pointer;
  }
  .search:hover { background: var(--accent); color: var(--accent-fg); }
  .search .keys { margin-left: auto; display: inline-flex; gap: 2px; }
  .links { display: flex; gap: 1.1rem; margin-left: auto; }
  .links a {
    font-size: 0.875rem;
    color: var(--muted-fg);
    text-decoration: none;
  }
  .links a:hover { color: var(--fg); }
  .end { display: flex; align-items: center; gap: 0.25rem; }
  .links:empty + .end { margin-left: auto; }
  .menu, .search-sm { display: none; }

  @media (max-width: 1000px) {
    .links { display: none; }
    .end { margin-left: auto; }
    .search { display: none; }
    .search-sm { display: inline-flex; }
  }
  @media (max-width: 800px) {
    .menu { display: inline-flex; }
    .home { display: none; }
  }
  .project-logo { width: 22px; height: 22px; object-fit: contain; border-radius: 4px; }
  @media print {
    .nav { display: none; }
  }
  .version {
    flex: none;
    height: 1.75rem;
    padding: 0 0.4rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--card);
    color: var(--fg);
    font: inherit;
    font-size: 0.8rem;
    cursor: pointer;
  }
  .version:focus-visible { outline: 2px solid var(--ring); outline-offset: 2px; }
</style>
