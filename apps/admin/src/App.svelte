<script lang="ts">
  import { api, session, SITE_URL, API_URL } from './lib/api.svelte'
  import { router, href } from './lib/router.svelte'
  import Toasts from './lib/Toasts.svelte'
  import Login from './pages/Login.svelte'
  import Dashboard from './pages/Dashboard.svelte'
  import Projects from './pages/Projects.svelte'
  import Project from './pages/Project.svelte'
  import PageEditor from './pages/PageEditor.svelte'
  import Tokens from './pages/Tokens.svelte'

  let checking = $state(false)

  // Resolve the user behind a stored token once per token.
  let checkedToken: string | null = null
  $effect(() => {
    const token = session.token
    if (!token || session.user || checkedToken === token) return
    checkedToken = token
    checking = true
    api
      .me()
      .then((r) => (session.user = r.user))
      .catch(() => {})
      .finally(() => (checking = false))
  })

  async function logout() {
    try {
      await api.logout()
    } catch {
      // Token may already be gone; sign out locally anyway.
    }
    session.set(null)
  }

  const section = $derived(
    router.route.name === 'tokens'
      ? 'tokens'
      : router.route.name === 'dashboard'
        ? 'dashboard'
        : 'projects',
  )
</script>

{#if !session.token}
  <Login />
{:else}
  <div class="shell">
    <aside>
      <a class="brand" href={href('/')}>
        <img src="/favicon.svg" alt="" width="22" height="22" />
        <span>jevidocs</span>
        <span class="badge">admin</span>
      </a>
      <nav>
        <a href={href('/')} aria-current={section === 'dashboard' ? 'page' : undefined}>
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 12l9-8 9 8M5 10v10h5v-6h4v6h5V10" /></svg>
          Dashboard
        </a>
        <a href={href('/projects')} aria-current={section === 'projects' ? 'page' : undefined}>
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 5h6l2 2h8v12H4z" /></svg>
          Projects
        </a>
        <a href={href('/tokens')} aria-current={section === 'tokens' ? 'page' : undefined}>
          <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="8" cy="15" r="4" /><path d="M11 12l9-9M17 6l3 3M15 8l2 2" /></svg>
          API tokens
        </a>
      </nav>
      <div class="links">
        <a href={SITE_URL} target="_blank" rel="noreferrer">Docs site ↗</a>
        <a href={API_URL + '/health'} target="_blank" rel="noreferrer">API status ↗</a>
      </div>
      <div class="user">
        <div class="who">
          {#if session.user}
            <strong>{session.user.name}</strong>
            <span class="muted">{session.user.email}</span>
          {:else if checking}
            <span class="muted">Loading…</span>
          {/if}
        </div>
        <button class="btn sm ghost" onclick={logout}>Sign out</button>
      </div>
    </aside>

    <main>
      {#if router.route.name === 'dashboard'}
        <Dashboard />
      {:else if router.route.name === 'projects'}
        <Projects />
      {:else if router.route.name === 'project'}
        {#key router.route.slug}
          <Project slug={router.route.slug} />
        {/key}
      {:else if router.route.name === 'page'}
        {#key `${router.route.slug}/${router.route.id}`}
          <PageEditor project={router.route.slug} id={router.route.id} />
        {/key}
      {:else if router.route.name === 'tokens'}
        <Tokens />
      {:else}
        <div class="empty">
          <h1>Not found</h1>
          <p><a href={href('/')}>Back to the dashboard</a></p>
        </div>
      {/if}
    </main>
  </div>
{/if}

<Toasts />

<style>
  .shell {
    display: grid;
    grid-template-columns: 15rem 1fr;
    min-height: 100vh;
  }
  aside {
    position: sticky;
    top: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    padding: 1rem 0.75rem;
    border-right: 1px solid var(--border);
    background: var(--surface);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 650;
    font-size: 1rem;
    text-decoration: none;
    padding: 0.25rem 0.5rem;
  }
  nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  nav a {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.45rem 0.6rem;
    border-radius: 8px;
    color: var(--muted);
    text-decoration: none;
    font-weight: 500;
  }
  nav a:hover {
    background: var(--surface-2);
    color: var(--text);
  }
  nav a[aria-current='page'] {
    background: var(--surface-2);
    color: var(--text);
  }
  nav svg {
    width: 16px;
    height: 16px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .links {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 0 0.6rem;
    font-size: 0.85rem;
  }
  .links a {
    color: var(--muted);
    text-decoration: none;
  }
  .links a:hover {
    color: var(--text);
  }
  .user {
    margin-top: auto;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 0.5rem 0;
    border-top: 1px solid var(--border);
  }
  .who {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
    font-size: 0.82rem;
  }
  .who span,
  .who strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  main {
    padding: 2rem clamp(1rem, 3vw, 2.5rem);
    min-width: 0;
  }
  @media (max-width: 760px) {
    .shell {
      grid-template-columns: 1fr;
    }
    aside {
      position: static;
      height: auto;
      flex-direction: row;
      flex-wrap: wrap;
      align-items: center;
      border-right: none;
      border-bottom: 1px solid var(--border);
    }
    nav {
      flex-direction: row;
    }
    .links {
      display: none;
    }
    .user {
      margin: 0 0 0 auto;
      border: none;
      padding: 0;
    }
  }
</style>
