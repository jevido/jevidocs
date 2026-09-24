<script lang="ts">
  import { api, type Project } from '../lib/api'
  import Logo from '../lib/Logo.svelte'
  import ThemeToggle from '../lib/ThemeToggle.svelte'
  import { readerAuth } from '../lib/reader-auth.svelte'
  import ReaderAccount from './ReaderAccount.svelte'

  let projects = $state<Project[] | null>(null)
  let error = $state('')

  $effect(() => {
    // Signed-in readers also see the private projects they may read.
    void readerAuth.version
    const ctrl = new AbortController()
    api
      .projects(ctrl.signal)
      .then((ps) => (projects = ps))
      .catch((e) => {
        if (!ctrl.signal.aborted) error = e instanceof Error ? e.message : String(e)
      })
    return () => ctrl.abort()
  })

  const href = (p: Project) => (p.slug === 'jevidocs' ? '/docs' : `/p/${p.slug}`)

  $effect(() => {
    document.title = 'Projects — jevidocs'
  })
</script>

<a class="skip-link" href="#content">Skip to content</a>
<header class="bar">
  <a class="brand" href="/" aria-label="jevidocs home"><Logo /></a>
  <nav>
    <a href="/docs">Docs</a>
    <a href="https://admin.jevidocs.jevido.app">Admin</a>
    <ThemeToggle />
    <ReaderAccount />
  </nav>
</header>

<main id="content" tabindex="-1">
  <h1>Projects</h1>
  <p class="lead">
    Documentation sites hosted on jevidocs.{#if !readerAuth.user}
      Private projects appear here once you sign in.{/if}
  </p>

  {#if error}
    <p class="error">Could not load projects: {error}</p>
  {:else if projects === null}
    <div class="grid">
      {#each [1, 2, 3] as i (i)}<div class="card skeleton"></div>{/each}
    </div>
  {:else if projects.length === 0}
    <p class="lead">No public projects yet.</p>
  {:else}
    <div class="grid">
      {#each projects as p (p.slug)}
        <a class="card" href={href(p)}>
          <span class="name">{p.name}{#if !p.public}<span class="private">Private</span>{/if}</span>
          {#if p.description}<span class="desc">{p.description}</span>{/if}
          <span class="path">{href(p)}</span>
        </a>
      {/each}
    </div>
  {/if}
</main>

<style>
  .bar {
    height: var(--nav-h);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 1.25rem;
    border-bottom: 1px solid var(--border);
  }
  .brand { color: var(--fg); text-decoration: none; }
  nav { display: flex; align-items: center; gap: 1rem; }
  nav a { color: var(--muted-fg); text-decoration: none; font-size: 0.9rem; }
  nav a:hover { color: var(--fg); }
  main { max-width: 60rem; margin: 0 auto; padding: 3rem 1.25rem; }
  h1 { font-size: 2rem; margin: 0 0 0.25rem; letter-spacing: -0.02em; }
  .lead { color: var(--muted-fg); margin: 0 0 2rem; }
  .error { color: var(--error); }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr)); gap: 0.75rem; }
  .card {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
    color: var(--fg);
    text-decoration: none;
    min-height: 6.5rem;
  }
  a.card:hover { background: var(--accent); }
  .name { font-weight: 600; }
  .desc { color: var(--muted-fg); font-size: 0.875rem; line-height: 1.5; }
  .path { margin-top: auto; font-family: var(--font-mono); font-size: 0.75rem; color: var(--muted-fg); }
  .skeleton { animation: pulse 1.2s ease-in-out infinite; }
  @keyframes pulse { 50% { opacity: 0.5; } }
  .private {
    margin-left: 0.5rem;
    font-size: 0.7rem;
    font-weight: 500;
    padding: 0.1rem 0.45rem;
    border-radius: 999px;
    vertical-align: middle;
    border: 1px solid color-mix(in oklab, var(--warn) 45%, var(--border));
    color: var(--warn);
  }
</style>
