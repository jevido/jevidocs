<script lang="ts">
  import { tick, untrack } from 'svelte'
  import { api, setLocale, ApiError, type Page, type ProjectWithTree } from '../lib/api'
  import { enhance, onContentClick, onContentKeydown } from '../lib/enhance'
  import Icon from '../lib/Icon.svelte'
  import { interceptLinks, router, scrollToHash } from '../lib/router.svelte'
  import Navbar from './Navbar.svelte'
  import Feedback from './Feedback.svelte'
  import PageActions from './PageActions.svelte'
  import SearchDialog from './SearchDialog.svelte'
  import Toc from './Toc.svelte'
  import TreeItems from './TreeItems.svelte'
  import RootToggle from './RootToggle.svelte'
  import ReaderExtras from './ReaderExtras.svelte'
  import MobileToc from './MobileToc.svelte'
  import { applyAccent, setPageMeta } from '../lib/project-theme'
  import { sidebar } from '../lib/sidebar.svelte'
  import { containsSlug } from '../lib/tree'
  import type { TreeNode } from '../lib/api'

  interceptLinks()

  const route = $derived(router.route)


  let project = $state<ProjectWithTree | null>(null)
  let projectError = $state<string | null>(null)

  // Root folders (fumadocs' sidebar tabs): the sidebar shows the tree of the
  // root that holds the current page, or everything outside any root.
  type RootFolder = Extract<TreeNode, { type: 'folder' }>
  const allNodes = $derived<TreeNode[]>(project?.tree?.children ?? [])
  const roots = $derived(allNodes.filter((n): n is RootFolder => n.type === 'folder' && !!n.root))
  const outside = $derived(allNodes.filter((n) => !(n.type === 'folder' && n.root)))
  const activeRoot = $derived(roots.find((r) => containsSlug(r, route.slug)) ?? null)
  const sidebarNodes = $derived<TreeNode[]>(
    activeRoot
      ? [...(activeRoot.index ? [{ ...activeRoot.index, name: 'Overview' }] : []), ...activeRoot.children]
      : outside,
  )

  let page = $state<Page | null>(null)
  let status = $state<'loading' | 'ready' | 'notfound' | 'error'>('loading')
  let errorMessage = $state('')

  let searchOpen = $state(false)
  let drawerOpen = $state(false)
  let content = $state<HTMLElement>()

  // Tell the router which leading URL segments are languages of this project.
  // Only on a real change: a new array would re-derive the route and refetch.
  $effect(() => {
    const next = project && project.slug === route.project ? (project.locales ?? []).slice(1) : []
    if (next.join(',') !== untrack(() => router.locales).join(',')) router.locales = next
  })

  $effect(() => {
    const slug = route.project
    setLocale(route.locale)
    projectError = null
    if (!slug) {
      project = null
      projectError = 'No project in this URL. Try /docs or /p/<project>.'
      return
    }
    const ctrl = new AbortController()
    api
      .project(slug, ctrl.signal)
      .then((p) => (project = p))
      .catch((e: unknown) => {
        if ((e as Error).name === 'AbortError') return
        project = null
        projectError =
          e instanceof ApiError && e.status === 404 ? `Project “${slug}” does not exist.` : 'Could not reach the docs API.'
      })
    return () => ctrl.abort()
  })

  $effect(() => {
    const { project: p, slug, locale } = route
    if (!p) return
    // A leading "nl/" may be a language: wait for the project to say so.
    if (/^[a-z]{2,3}(-[a-z0-9]+)?(\/|$)/.test(slug) && project?.slug !== p) return
    setLocale(locale)
    const ctrl = new AbortController()
    status = 'loading'
    drawerOpen = false
    api
      .page(p, slug, ctrl.signal)
      .then(async (data) => {
        page = data
        status = 'ready'
        api.trackView(p, data.slug)
        await tick()
        const hash = untrack(() => router.hash)
        scrollToHash(hash)
      })
      .catch((e: unknown) => {
        if ((e as Error).name === 'AbortError') return
        page = null
        if (e instanceof ApiError && e.status === 404) status = 'notfound'
        else {
          status = 'error'
          errorMessage = (e as Error).message
        }
      })
    return () => ctrl.abort()
  })

  $effect(() => {
    void page?.html
    if (content) enhance(content)
  })

  $effect(() => {
    const name = project?.name ?? 'Docs'
    document.title =
      status === 'ready' && page ? `${page.title} | ${name}` : status === 'notfound' ? `Not found | ${name}` : name
  })

  $effect(() => applyAccent(project?.accent))
  $effect(() => {
    if (status === 'ready' && page) {
      setPageMeta(page.title, page.description || project?.description || '', location.origin + router.href(page.slug))
    }
  })

  const crumbs = $derived((page?.breadcrumbs ?? []).slice(0, -1))
  // ~220 words a minute; code and markup count too, close enough for a hint.
  const readingMinutes = $derived(
    page?.markdown ? Math.max(1, Math.round(page.markdown.split(/\s+/).filter(Boolean).length / 220)) : 0,
  )

  const updated = $derived.by(() => {
    if (!page?.updated_at) return ''
    const d = new Date(page.updated_at)
    return isNaN(d.getTime()) ? '' : d.toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' })
  })
  const toc = $derived(page?.toc ?? [])
</script>

<a class="skip-link" href="#content">Skip to content</a>
{#if project?.banner}
  <div class="banner" role="note">{project.banner}</div>
{/if}
<Navbar {project} onsearch={() => (searchOpen = true)} onmenu={() => (drawerOpen = true)} />

{#if route.project}
  <SearchDialog bind:open={searchOpen} project={route.project} ask={project?.ask ?? false} />
{/if}

{#if status === 'loading' && page}
  <div class="progress" aria-hidden="true"></div>
{/if}

<ReaderExtras previous={page?.previous?.slug} next={page?.next?.slug} onsearch={() => (searchOpen = true)} />

<div class="layout" class:collapsed={sidebar.collapsed}>
  {#if drawerOpen}
    <div class="scrim" role="presentation" onclick={() => (drawerOpen = false)}></div>
  {/if}
  <aside class="sidebar" class:open={drawerOpen} aria-label="Documentation navigation">
    <div class="sidebar-head">
      <span>{project?.name ?? ''}</span>
      <button class="icon-btn" type="button" aria-label="Close navigation" onclick={() => (drawerOpen = false)}>
        <Icon name="x" size={18} />
      </button>
    </div>
    <div class="sidebar-scroll">
      {#if project}
        {#if roots.length > 0}
          <RootToggle
            projectName={project.name}
            {roots}
            active={activeRoot}
            rest={outside}
            onnavigate={() => (drawerOpen = false)} />
        {/if}
        <TreeItems nodes={sidebarNodes} onnavigate={() => (drawerOpen = false)} />
      {:else if !projectError}
        {#each Array(8) as _, i (i)}
          <div class="sk sk-item" style:width="{55 + ((i * 37) % 40)}%"></div>
        {/each}
      {/if}
    </div>
    {#if project}
      <div class="sidebar-foot">
        <a href={api.llmsUrl(route.project)} target="_blank" rel="noreferrer">llms.txt</a>
        <a href={api.llmsFullUrl(route.project)} target="_blank" rel="noreferrer">llms-full.txt</a>
        <button class="collapse" type="button" aria-label="Collapse sidebar" title="Collapse sidebar" onclick={() => sidebar.toggle(true)}>
          <Icon name="chevronLeft" size={15} />
        </button>
      </div>
    {/if}
  </aside>

  <main class="main" id="content" tabindex="-1">
    {#if projectError}
      <div class="state">
        <p class="code">404</p>
        <h1>Project not found</h1>
        <p>{projectError}</p>
        <a class="btn btn-secondary" href="/docs">Go to the jevidocs docs</a>
      </div>
    {:else if status === 'notfound'}
      <div class="state">
        <p class="code">404</p>
        <h1>Page not found</h1>
        <p>The page <code>/{route.slug}</code> does not exist in {project?.name ?? route.project}.</p>
        <a class="btn btn-secondary" href={router.href('')}>Back to the introduction</a>
      </div>
    {:else if status === 'error'}
      <div class="state">
        <h1>Something went wrong</h1>
        <p>{errorMessage || 'The page could not be loaded.'}</p>
        <button class="btn btn-secondary" type="button" onclick={() => location.reload()}>
          Try again
        </button>
      </div>
    {:else if !page}
      <div class="skeleton" aria-busy="true" aria-label="Loading">
        <div class="sk" style="width: 40%; height: 2.2rem"></div>
        <div class="sk" style="width: 70%; height: 1.1rem; margin-bottom: 2rem"></div>
        {#each Array(6) as _, i (i)}
          <div class="sk" style:width="{70 + ((i * 13) % 30)}%"></div>
        {/each}
      </div>
    {:else}
      <article class:dim={status === 'loading'}>
        {#if crumbs.length}
          <nav class="crumbs" aria-label="Breadcrumb">
            {#each crumbs as c, i (i)}
              {#if i > 0}<Icon name="chevronRight" size={13} />{/if}
              {#if c.slug !== undefined}
                <a href={router.href(c.slug)}>{c.name}</a>
              {:else}
                <span>{c.name}</span>
              {/if}
            {/each}
          </nav>
        {/if}
        <h1 class="title">{page.title}</h1>
        {#if page.description}<p class="description">{page.description}</p>{/if}
        <PageActions markdown={page.markdown} markdownUrl={api.markdownUrl(route.project, page.slug)} />
        <div class="divider"></div>
        <MobileToc items={toc} />

        {#if page.draft}
          <p class="draft-note" role="note">Draft preview: this page is not published yet.</p>
        {/if}
        {#if page.fallback}
          <p class="fallback-note" role="note">This page is not translated yet; showing the original.</p>
        {/if}
        <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
        <div class="prose" bind:this={content} onclick={onContentClick} onkeydown={onContentKeydown}>
          {@html page.html}
        </div>

        <footer class="page-foot">
          <Feedback project={route.project} slug={page.slug} />
          <div class="meta-row">
            {#if page.edit_url}
              <a class="edit" href={page.edit_url} target="_blank" rel="noreferrer">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M12 20h9" /><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" /></svg>
                Edit this page
              </a>
            {/if}
            {#if updated}<p class="updated">
                Last updated on {updated}{#if readingMinutes}{' · '}{readingMinutes} min read{/if}
              </p>{/if}
          </div>
          {#if page.previous || page.next}
            <div class="pager">
              {#if page.previous}
                <a class="card prev" href={router.href(page.previous.slug)}>
                  <span class="dir"><Icon name="chevronLeft" size={14} /> Previous</span>
                  <span class="name">{page.previous.title}</span>
                </a>
              {:else}
                <span></span>
              {/if}
              {#if page.next}
                <a class="card next" href={router.href(page.next.slug)}>
                  <span class="dir">Next <Icon name="chevronRight" size={14} /></span>
                  <span class="name">{page.next.title}</span>
                </a>
              {/if}
            </div>
          {/if}
        </footer>
      </article>
    {/if}
  </main>

  <aside class="toc-col">
    {#if page && status !== 'notfound'}
      <div class="toc-sticky">
        <Toc items={toc} {content} />
      </div>
    {/if}
  </aside>
</div>

<style>
  .progress {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 70;
    height: 2px;
    width: 40%;
    background: var(--brand);
    animation: load 1s ease-in-out infinite;
  }
  @keyframes load {
    from { transform: translateX(-100%); }
    to { transform: translateX(260%); }
  }
  .layout {
    display: grid;
    grid-template-columns: var(--sidebar-w) minmax(0, 1fr) var(--toc-w);
    max-width: 1400px;
    margin: 0 auto;
  }
  .sidebar {
    position: sticky;
    top: var(--nav-h);
    height: calc(100vh - var(--nav-h));
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border);
  }
  .sidebar-head { display: none; }
  .sidebar-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 1.25rem 0.75rem;
    scrollbar-width: thin;
  }
  .sidebar-foot {
    display: flex;
    gap: 1rem;
    padding: 0.75rem 1.25rem;
    border-top: 1px solid var(--border);
    font-size: 0.75rem;
  }
  .sidebar-foot a { color: var(--muted-fg); text-decoration: none; font-family: var(--font-mono); }
  .sidebar-foot a:hover { color: var(--fg); }
  .sidebar-foot .collapse {
    margin-left: auto;
    display: grid;
    place-items: center;
    border: 0;
    background: none;
    color: var(--muted-fg);
    cursor: pointer;
    padding: 0;
  }
  .sidebar-foot .collapse:hover { color: var(--fg); }
  @media (min-width: 801px) {
    .layout.collapsed { grid-template-columns: minmax(0, 1fr) var(--toc-w); }
    .layout.collapsed .sidebar { display: none; }
  }
  @media (min-width: 801px) and (max-width: 1200px) {
    .layout.collapsed { grid-template-columns: minmax(0, 1fr); }
  }
  @media (max-width: 800px) {
    .sidebar-foot .collapse { display: none; }
  }
  @media print {
    .layout { display: block; }
    .sidebar,
    .toc-col,
    .banner,
    .progress,
    .page-foot :global(.feedback),
    .pager { display: none !important; }
    .main { padding: 0; }
  }

  .main {
    min-width: 0;
    padding: 2.5rem 2.5rem 4rem;
  }
  article { max-width: 860px; margin: 0 auto; transition: opacity 0.2s; }
  article.dim { opacity: 0.6; }
  .crumbs {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.3rem;
    margin-bottom: 0.75rem;
    font-size: 0.85rem;
    color: var(--muted-fg);
  }
  .crumbs a { text-decoration: none; }
  .crumbs a:hover { color: var(--fg); }
  .title {
    margin: 0 0 0.5rem;
    font-size: 1.9rem;
    font-weight: 600;
    letter-spacing: -0.02em;
    line-height: 1.2;
  }
  .description {
    margin: 0 0 1.25rem;
    font-size: 1.1rem;
    color: var(--muted-fg);
  }
  .divider { height: 1px; background: var(--border); margin: 1.5rem 0 0.5rem; }

  .page-foot { margin-top: 3.5rem; }
  .updated { font-size: 0.85rem; color: var(--muted-fg); margin: 0 0 1.25rem; }
  .pager { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .card {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
    text-decoration: none;
    transition: background 0.15s;
  }
  .card:hover { background: var(--accent); }
  .card.next { text-align: right; align-items: flex-end; grid-column: 2; }
  .dir { display: inline-flex; align-items: center; gap: 0.2rem; font-size: 0.8rem; color: var(--muted-fg); }
  .name { font-weight: 500; font-size: 0.95rem; }

  .toc-col { padding: 2.5rem 1.25rem 0 0; }
  .toc-sticky {
    position: sticky;
    top: calc(var(--nav-h) + 2.5rem);
    max-height: calc(100vh - var(--nav-h) - 4rem);
    overflow-y: auto;
    scrollbar-width: thin;
  }

  .state {
    max-width: 560px;
    margin: 4rem auto;
    text-align: center;
  }
  .state .code {
    font-family: var(--font-mono);
    font-size: 0.85rem;
    color: var(--brand);
    margin: 0;
  }
  .state h1 { margin: 0.5rem 0; font-size: 1.8rem; letter-spacing: -0.02em; }
  .state p { color: var(--muted-fg); }

  .sk {
    height: 0.95rem;
    margin: 0.75rem 0;
    border-radius: 6px;
    background: linear-gradient(90deg, var(--muted), var(--accent), var(--muted));
    background-size: 200% 100%;
    animation: shimmer 1.4s linear infinite;
  }
  .sk-item { margin: 0.9rem 0.5rem; height: 0.8rem; }
  .skeleton { max-width: 860px; margin: 0 auto; }
  @keyframes shimmer { to { background-position: -200% 0; } }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 49;
    background: var(--overlay);
  }

  @media (max-width: 1200px) {
    .layout { grid-template-columns: var(--sidebar-w) minmax(0, 1fr); }
    .toc-col { display: none; }
  }
  @media (max-width: 800px) {
    .layout { grid-template-columns: minmax(0, 1fr); }
    .main { padding: 1.5rem 1rem 3rem; }
    .sidebar {
      position: fixed;
      z-index: 50;
      top: 0;
      left: 0;
      bottom: 0;
      height: 100vh;
      width: min(85vw, 320px);
      background: var(--bg);
      transform: translateX(-100%);
      transition: transform 0.2s ease;
      visibility: hidden;
    }
    .sidebar.open { transform: none; visibility: visible; box-shadow: var(--shadow); }
    .sidebar-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.6rem 0.75rem 0.6rem 1.25rem;
      border-bottom: 1px solid var(--border);
      font-weight: 600;
    }
    .pager { grid-template-columns: 1fr; }
    .card.next { grid-column: 1; }
  }
  .banner {
    padding: 0.5rem 1rem;
    text-align: center;
    font-size: 0.85rem;
    background: color-mix(in oklab, var(--brand) 14%, var(--bg));
    border-bottom: 1px solid var(--border);
  }
  .meta-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem 1rem;
  }
  .meta-row .updated { margin: 0; }
  .edit {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.85rem;
    color: var(--muted-fg);
    text-decoration: none;
  }
  .edit:hover { color: var(--fg); }
  .draft-note {
    margin: 0 0 1rem;
    padding: 0.5rem 0.75rem;
    border: 1px dashed var(--warn);
    border-radius: var(--radius);
    font-size: 0.85rem;
    color: var(--fg);
  }
  .fallback-note {
    margin: 0 0 1rem;
    padding: 0.5rem 0.75rem;
    border: 1px dashed var(--border);
    border-radius: var(--radius);
    font-size: 0.85rem;
    color: var(--muted-fg);
  }
</style>
