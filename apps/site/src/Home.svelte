<script lang="ts">
  import Icon from './lib/Icon.svelte'
  import Logo from './lib/Logo.svelte'
  import ThemeToggle from './lib/ThemeToggle.svelte'
  import { copyText } from './lib/enhance'

  const ADMIN_URL = 'https://admin.jevidocs.jevido.app'
  const API_URL = 'https://api.jevidocs.jevido.app'
  const GITHUB_URL = 'https://github.com/jevido/jevidocs'

  type Tab = 'content' | 'api' | 'mcp'
  let tab = $state<Tab>('content')
  let copied = $state(false)

  const install = 'git clone https://github.com/jevido/jevidocs && task api:dev'

  const features = [
    { icon: 'pen', title: 'Write in Markdown', text: 'CommonMark + GFM with callouts, tabs, steps, cards, accordions and file trees. Rendered by Go, not in the browser.' },
    { icon: 'layout', title: 'A docs UI that feels right', text: 'Sidebar page tree, table of contents with scrollspy, breadcrumbs, prev/next, dark mode. Built with Svelte 5 runes.' },
    { icon: 'search', title: 'Search built in', text: 'Full-text search on PostgreSQL across pages and headings, with highlighted snippets and ⌘K everywhere.' },
    { icon: 'settings', title: 'Admin included', text: 'Manage projects and pages in a web admin with a live preview. No rebuilds, no redeploys, just publish.' },
    { icon: 'bot', title: 'MCP for agents', text: 'A hosted MCP server lets LLMs list, read, search and even write your docs with a scoped token.' },
    { icon: 'sparkles', title: 'AI-ready output', text: 'llms.txt, llms-full.txt and a Markdown view of every page. Copy a page or open it in ChatGPT or Claude.' },
    { icon: 'database', title: 'PostgreSQL + Goravel', text: 'A plain Goravel JSON API over Postgres. Migrations on start, health checks, and one binary to deploy.' },
    { icon: 'zap', title: 'Fast by default', text: 'Server-side rendering of Markdown and syntax highlighting with Chroma. The client only styles and routes.' },
    { icon: 'package', title: 'Multi-project', text: 'Host as many documentation sites as you like; each gets its own tree, search index and llms.txt.' },
    { icon: 'code', title: 'OpenAPI reference', text: 'Import an OpenAPI 3 spec and get a page per operation: parameters, bodies, responses and curl/JS/Go examples.' },
    { icon: 'text', title: 'Docs as code', text: 'Keep Markdown in git and run jevidocs push, or pull a project down to files. The CLI talks to the same API.' },
    { icon: 'book', title: 'History and feedback', text: 'Every edit keeps a revision you can diff and restore, and readers tell you which pages help.' },
  ]

  const stack = ['Go 1.27', 'Goravel', 'Svelte 5', 'Vite', 'PostgreSQL 18', 'MCP', 'Coolify']
</script>

<div class="page">
  <a class="skip-link" href="#content">Skip to content</a>
  <header class="nav">
    <div class="inner">
      <a class="brand" href="/"><Logo /> jevidocs</a>
      <nav class="links">
        <a href="/docs">Docs</a>
        <a href="/docs/components">Components</a>
        <a href="/docs/mcp">MCP</a>
        <a href={ADMIN_URL}>Admin</a>
      </nav>
      <div class="end">
        <a class="icon-btn" href={GITHUB_URL} aria-label="GitHub" target="_blank" rel="noreferrer"><Icon name="github" size={18} /></a>
        <ThemeToggle />
      </div>
    </div>
  </header>

  <main id="content" tabindex="-1">
    <section class="hero">
      <div class="glow" aria-hidden="true"></div>
      <a class="badge" href="/docs/mcp">
        <span class="dot"></span> Now with a hosted MCP server <Icon name="arrowRight" size={13} />
      </a>
      <h1>Build excellent<br /><span class="grad">documentation sites</span></h1>
      <p class="lead">
        jevidocs is a documentation framework on Go, Goravel, Svelte 5 and PostgreSQL. Write Markdown, manage it in an
        admin, serve it through an API, and let AI agents read it over MCP.
      </p>
      <div class="cta">
        <a class="btn btn-primary" href="/docs">Getting started <Icon name="arrowRight" size={15} /></a>
        <a class="btn btn-secondary" href={ADMIN_URL}>Open the admin</a>
      </div>
      <button
        class="install"
        type="button"
        onclick={async () => {
          if (await copyText(install)) {
            copied = true
            setTimeout(() => (copied = false), 1500)
          }
        }}>
        <span class="prompt">$</span>
        <code>{install}</code>
        <Icon name={copied ? 'check' : 'copy'} size={14} />
      </button>
    </section>

    <section class="preview">
      <div class="window">
        <div class="bar">
          <span class="lights"><i></i><i></i><i></i></span>
          <div class="tabs" role="tablist">
            <button role="tab" aria-selected={tab === 'content'} class:on={tab === 'content'} onclick={() => (tab = 'content')}>content/index.md</button>
            <button role="tab" aria-selected={tab === 'api'} class:on={tab === 'api'} onclick={() => (tab = 'api')}>REST API</button>
            <button role="tab" aria-selected={tab === 'mcp'} class:on={tab === 'mcp'} onclick={() => (tab = 'mcp')}>MCP</button>
          </div>
        </div>
        {#if tab === 'content'}
          <pre><code><span class="c">---</span>
<span class="k">title</span>: Introduction
<span class="k">description</span>: Everything you need to ship docs.
<span class="c">---</span>

<span class="h">## Quick start</span>

<span class="t">&lt;Callout</span> <span class="k">type</span>=<span class="s">"info"</span><span class="t">&gt;</span>
Pages are Markdown with **components**.
<span class="t">&lt;/Callout&gt;</span>

<span class="t">&lt;Tabs</span> <span class="k">items</span>=<span class="s">"bun,go"</span><span class="t">&gt;</span>
<span class="t">&lt;Tab</span> <span class="k">value</span>=<span class="s">"bun"</span><span class="t">&gt;</span>`bun install`<span class="t">&lt;/Tab&gt;</span>
<span class="t">&lt;Tab</span> <span class="k">value</span>=<span class="s">"go"</span><span class="t">&gt;</span>`go run .`<span class="t">&lt;/Tab&gt;</span>
<span class="t">&lt;/Tabs&gt;</span></code></pre>
        {:else if tab === 'api'}
          <pre><code><span class="c"># the page tree of a project</span>
<span class="k">curl</span> {API_URL}/api/projects/jevidocs

<span class="c"># one rendered page: html, toc, breadcrumbs, prev/next</span>
<span class="k">curl</span> <span class="s">"{API_URL}/api/projects/jevidocs/page?slug=components"</span>

<span class="c"># full-text search across pages and headings</span>
<span class="k">curl</span> <span class="s">"{API_URL}/api/projects/jevidocs/search?q=tabs"</span>

<span class="c"># everything, for your LLM</span>
<span class="k">curl</span> {API_URL}/api/projects/jevidocs/llms-full.txt</code></pre>
        {:else}
          <pre><code><span class="c">// claude_desktop_config.json / .mcp.json</span>
{'{'}
  <span class="s">"mcpServers"</span>: {'{'}
    <span class="s">"jevidocs"</span>: {'{'}
      <span class="s">"type"</span>: <span class="s">"http"</span>,
      <span class="s">"url"</span>: <span class="s">"{API_URL}/mcp"</span>
    {'}'}
  {'}'}
{'}'}

<span class="c">// tools: list_projects, get_page_tree, read_page,</span>
<span class="c">//        search_docs, create_page, update_page, delete_page</span></code></pre>
        {/if}
      </div>
    </section>

    <section class="stack" aria-label="Stack">
      {#each stack as s (s)}<span>{s}</span>{/each}
    </section>

    <section class="features">
      <h2>Everything a docs site needs</h2>
      <p class="sub">The good parts of fumadocs, rebuilt on a Go backend with a Svelte front.</p>
      <div class="grid">
        {#each features as f (f.title)}
          <article>
            <span class="ficon"><Icon name={f.icon} size={18} /></span>
            <h3>{f.title}</h3>
            <p>{f.text}</p>
          </article>
        {/each}
      </div>
    </section>

    <section class="arch">
      <h2>Three parts, one repository</h2>
      <div class="cols">
        <a class="part" href="/docs">
          <span class="ficon"><Icon name="book" size={18} /></span>
          <h3>Site</h3>
          <p>This site and the docs reader at <code>/docs</code>. Svelte 5 + Vite, no SvelteKit.</p>
          <span class="more">Read the docs <Icon name="arrowRight" size={13} /></span>
        </a>
        <a class="part" href={ADMIN_URL}>
          <span class="ficon"><Icon name="settings" size={18} /></span>
          <h3>Admin</h3>
          <p>Projects, pages, live preview and API tokens for MCP writes.</p>
          <span class="more">Open the admin <Icon name="arrowRight" size={13} /></span>
        </a>
        <a class="part" href="{API_URL}/health">
          <span class="ficon"><Icon name="terminal" size={18} /></span>
          <h3>API + MCP</h3>
          <p>Goravel over PostgreSQL: rendering, trees, search, llms.txt and <code>/mcp</code>.</p>
          <span class="more">Check health <Icon name="arrowRight" size={13} /></span>
        </a>
      </div>
    </section>

    <section class="final">
      <h2>Start documenting.</h2>
      <p>Your docs deserve better than a wiki.</p>
      <a class="btn btn-primary" href="/docs">Read the docs <Icon name="arrowRight" size={15} /></a>
    </section>
  </main>

  <footer>
    <div class="inner">
      <span class="brand small"><Logo size={18} /> jevidocs</span>
      <span>Inspired by <a href="https://fumadocs.dev" target="_blank" rel="noreferrer">fumadocs</a>. © 2026 jevido.</span>
    </div>
  </footer>
</div>

<style>
  .nav {
    position: sticky;
    top: 0;
    z-index: 20;
    border-bottom: 1px solid var(--border);
    background: color-mix(in oklab, var(--bg) 80%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
  .inner {
    display: flex;
    align-items: center;
    gap: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1.25rem;
    height: var(--nav-h);
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 600;
    text-decoration: none;
  }
  .links { display: flex; gap: 1.25rem; }
  .links a { color: var(--muted-fg); text-decoration: none; font-size: 0.875rem; }
  .links a:hover { color: var(--fg); }
  .end { margin-left: auto; display: flex; align-items: center; gap: 0.25rem; }

  main { max-width: 1200px; margin: 0 auto; padding: 0 1.25rem; }

  .hero {
    position: relative;
    text-align: center;
    padding: 6rem 0 3rem;
  }
  .glow {
    position: absolute;
    inset: -2rem 10% auto;
    height: 380px;
    z-index: -1;
    background:
      radial-gradient(40% 60% at 30% 40%, color-mix(in oklab, #a78bfa 35%, transparent), transparent 70%),
      radial-gradient(40% 60% at 70% 40%, color-mix(in oklab, #38bdf8 30%, transparent), transparent 70%);
    filter: blur(40px);
    opacity: 0.55;
  }
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.3rem 0.8rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--card);
    color: var(--muted-fg);
    font-size: 0.8rem;
    text-decoration: none;
  }
  .badge:hover { color: var(--fg); }
  .dot { width: 6px; height: 6px; border-radius: 50%; background: var(--success); box-shadow: 0 0 8px var(--success); }
  h1 {
    margin: 1.5rem 0 1.25rem;
    font-size: clamp(2.4rem, 7vw, 4.4rem);
    line-height: 1.05;
    letter-spacing: -0.04em;
    font-weight: 700;
  }
  .grad {
    background: linear-gradient(90deg, #a78bfa, #38bdf8);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }
  .lead {
    max-width: 40rem;
    margin: 0 auto 2rem;
    color: var(--muted-fg);
    font-size: 1.1rem;
  }
  .cta { display: flex; gap: 0.75rem; justify-content: center; flex-wrap: wrap; }
  .install {
    display: inline-flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 1.75rem;
    padding: 0.6rem 0.9rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
    color: var(--muted-fg);
    font-family: var(--font-mono);
    font-size: 0.8rem;
    cursor: pointer;
    max-width: 100%;
  }
  .install code { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--fg); }
  .install:hover { background: var(--accent); }
  .prompt { color: var(--brand); }

  .preview { display: flex; justify-content: center; padding: 1rem 0 3rem; }
  .window {
    width: min(760px, 100%);
    border: 1px solid var(--border);
    border-radius: 0.9rem;
    background: var(--card);
    box-shadow: var(--shadow);
    overflow: hidden;
  }
  .bar {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0 0.9rem;
    border-bottom: 1px solid var(--border);
    background: var(--muted);
  }
  .lights { display: inline-flex; gap: 6px; }
  .lights i { width: 10px; height: 10px; border-radius: 50%; background: var(--border); }
  .tabs { display: flex; gap: 0.2rem; overflow-x: auto; }
  .tabs button {
    padding: 0.65rem 0.7rem;
    border: 0;
    border-bottom: 2px solid transparent;
    background: none;
    color: var(--muted-fg);
    font-family: var(--font-mono);
    font-size: 0.78rem;
    cursor: pointer;
    white-space: nowrap;
  }
  .tabs button.on { color: var(--fg); border-bottom-color: var(--brand); }
  pre {
    margin: 0;
    padding: 1.25rem 1.4rem;
    min-height: 320px;
    overflow-x: auto;
    font-family: var(--font-mono);
    font-size: 0.82rem;
    line-height: 1.7;
    text-align: left;
  }
  pre .c { color: var(--muted-fg); }
  pre .k { color: #cf222e; }
  pre .s { color: #0a3069; }
  pre .t { color: #116329; }
  pre .h { color: #8250df; font-weight: 600; }
  :global(.dark) pre .k { color: #ff7b72; }
  :global(.dark) pre .s { color: #a5d6ff; }
  :global(.dark) pre .t { color: #7ee787; }
  :global(.dark) pre .h { color: #d2a8ff; }

  .stack {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.5rem 1.75rem;
    padding: 1rem 0 4rem;
    color: var(--muted-fg);
    font-family: var(--font-mono);
    font-size: 0.85rem;
  }

  h2 {
    margin: 0 0 0.5rem;
    font-size: clamp(1.6rem, 4vw, 2.2rem);
    letter-spacing: -0.03em;
    text-align: center;
  }
  .sub { text-align: center; color: var(--muted-fg); margin: 0 0 2.5rem; }
  .grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    border: 1px solid var(--border);
    border-radius: 1rem;
    overflow: hidden;
  }
  .grid article {
    padding: 1.6rem;
    border-right: 1px solid var(--border);
    border-bottom: 1px solid var(--border);
    background: var(--bg);
    transition: background 0.2s;
  }
  .grid article:hover { background: var(--card); }
  .grid article:nth-child(3n) { border-right: 0; }
  .grid article:nth-last-child(-n + 3) { border-bottom: 0; }
  .ficon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.2rem;
    height: 2.2rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--card);
    color: var(--brand);
    margin-bottom: 0.9rem;
  }
  h3 { margin: 0 0 0.35rem; font-size: 1rem; font-weight: 600; }
  article p, .part p { margin: 0; color: var(--muted-fg); font-size: 0.9rem; }

  .arch { padding: 5rem 0 0; }
  .cols { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; margin-top: 2rem; }
  .part {
    display: flex;
    flex-direction: column;
    padding: 1.5rem;
    border: 1px solid var(--border);
    border-radius: 1rem;
    background: var(--card);
    text-decoration: none;
    transition: border-color 0.2s, transform 0.2s;
  }
  .part:hover { border-color: color-mix(in oklab, var(--brand) 45%, var(--border)); transform: translateY(-2px); }
  .part code { font-family: var(--font-mono); font-size: 0.85em; }
  .more { margin-top: 1rem; display: inline-flex; align-items: center; gap: 0.3rem; font-size: 0.85rem; color: var(--brand); }

  .final { text-align: center; padding: 6rem 0; }
  .final p { color: var(--muted-fg); margin: 0 0 1.5rem; }

  footer { border-top: 1px solid var(--border); }
  footer .inner { justify-content: space-between; font-size: 0.85rem; color: var(--muted-fg); flex-wrap: wrap; height: auto; padding-block: 1.5rem; }
  footer a { color: var(--fg); }
  .small { font-size: 0.9rem; color: var(--fg); }

  @media (max-width: 900px) {
    .grid { grid-template-columns: 1fr 1fr; }
    .grid article:nth-child(3n) { border-right: 1px solid var(--border); }
    .grid article:nth-child(2n) { border-right: 0; }
    .grid article:nth-last-child(-n + 3) { border-bottom: 1px solid var(--border); }
    .grid article:last-child { border-bottom: 0; }
    .cols { grid-template-columns: 1fr; }
  }
  @media (max-width: 640px) {
    .links { display: none; }
    .hero { padding-top: 4rem; }
    .grid { grid-template-columns: 1fr; }
    .grid article { border-right: 0 !important; }
  }
</style>
