<script lang="ts">
  import { clients, type ApiReference } from '../../lib/openapi'
  import ApiClient from './ApiClient.svelte'
  import ApiModel from './ApiModel.svelte'
  import ApiOperation from './ApiOperation.svelte'
  import { ApiState, apiNav, setApi } from './state.svelte'

  // A Scalar-style reference: introduction with server, authentication and
  // client pickers, then every tag's operations in two columns (docs left,
  // request and response samples right), then the models.
  let { reference, project }: { reference: ApiReference; project: string } = $props()

  // The parent re-creates this component per page, so reading the props
  // once is enough.
  // svelte-ignore state_referenced_locally
  const api = setApi(new ApiState(reference, project))
  const ref = api.ref

  const scheme = $derived(ref.security_schemes.find((s) => s.key === api.scheme))
  const server = $derived(ref.servers[api.serverIndex])

  // Scrollspy for the sidebar: the last section whose top has passed the
  // navbar is the one being read.
  function spy(root: HTMLElement) {
    const sections = [...root.querySelectorAll<HTMLElement>('[data-op], [data-tag], [data-model]')]
    let frame = 0
    const update = () => {
      frame = 0
      const line = 56 + 80
      let current = ''
      for (const s of sections) {
        if (s.getBoundingClientRect().top <= line) current = s.id
        else break
      }
      apiNav.active = current
    }
    const onScroll = () => {
      if (!frame) frame = requestAnimationFrame(update)
    }
    update()
    addEventListener('scroll', onScroll, { passive: true })
    return () => {
      removeEventListener('scroll', onScroll)
      cancelAnimationFrame(frame)
      apiNav.active = ''
    }
  }
</script>

<div class="ref" {@attach spy}>
  <section class="intro">
    <div class="intro-main">
      <div class="badges">
        {#if ref.version}<span class="badge">v{ref.version}</span>{/if}
        <span class="badge">OAS {ref.openapi}</span>
      </div>
      <h1>{ref.title}</h1>
      {#if ref.description}<div class="prose">{@html ref.description}</div>{/if}
    </div>

    <div class="intro-side">
      <div class="card">
        <p class="card-title">Server</p>
        {#if ref.servers.length > 1}
          <select bind:value={api.serverIndex} aria-label="Server">
            {#each ref.servers as s, i (i)}<option value={i}>{s.url}{s.description ? ` (${s.description})` : ''}</option>{/each}
          </select>
        {:else if ref.servers.length === 1}
          <code class="url">{api.server}</code>
          {#if server?.description}<p class="hint">{server.description}</p>{/if}
        {:else}
          <input type="url" placeholder="https://api.example.com" bind:value={api.customServer} aria-label="Server URL" />
          <p class="hint">The spec lists no server; enter the base URL to use.</p>
        {/if}
        {#each server?.variables ?? [] as v (v.name)}
          <label class="field">
            <span>{v.name}</span>
            {#if v.enum?.length}
              <select bind:value={api.serverVars[v.name]}>
                {#each v.enum as e (e)}<option>{e}</option>{/each}
              </select>
            {:else}
              <input bind:value={api.serverVars[v.name]} placeholder={v.default} />
            {/if}
          </label>
        {/each}
      </div>

      {#if ref.security_schemes.length}
        <div class="card">
          <p class="card-title">Authentication</p>
          {#if ref.security_schemes.length > 1}
            <select bind:value={api.scheme} aria-label="Authentication scheme">
              {#each ref.security_schemes as s (s.key)}<option value={s.key}>{s.key}</option>{/each}
            </select>
          {/if}
          {#if scheme}
            {@const a = api.auth[scheme.key] ?? {}}
            {#if scheme.type === 'http' && scheme.scheme === 'basic'}
              <label class="field"><span>Username</span>
                <input value={a.username ?? ''} oninput={(e) => api.setAuth(scheme.key, 'username', e.currentTarget.value)} autocomplete="off" />
              </label>
              <label class="field"><span>Password</span>
                <input type="password" value={a.password ?? ''} oninput={(e) => api.setAuth(scheme.key, 'password', e.currentTarget.value)} autocomplete="off" />
              </label>
            {:else if scheme.type === 'apiKey'}
              <label class="field"><span>{scheme.name} <small>({scheme.in})</small></span>
                <input type="password" value={a.value ?? ''} placeholder="API key" oninput={(e) => api.setAuth(scheme.key, 'value', e.currentTarget.value)} autocomplete="off" />
              </label>
            {:else}
              <label class="field"><span>{scheme.type === 'http' ? `${scheme.scheme ?? 'bearer'} token` : 'Access token'}{scheme.bearer_format ? ` (${scheme.bearer_format})` : ''}</span>
                <input type="password" value={a.token ?? ''} placeholder="Token" oninput={(e) => api.setAuth(scheme.key, 'token', e.currentTarget.value)} autocomplete="off" />
              </label>
            {/if}
            {#if scheme.description}<div class="hint prose">{@html scheme.description}</div>{/if}
            <p class="hint">Kept in this tab only; used by the examples and Test Request.</p>
          {/if}
        </div>
      {/if}

      <div class="card">
        <p class="card-title">Client libraries</p>
        <div class="clients" role="radiogroup" aria-label="Code sample language">
          {#each clients as c (c.id)}
            <button type="button" role="radio" aria-checked={api.client === c.id} onclick={() => api.setClient(c.id)}>{c.label}</button>
          {/each}
        </div>
      </div>
    </div>
  </section>

  {#each ref.tags as tag (tag.id)}
    <section class="tag" id={tag.id} data-tag>
      <div class="tag-head">
        <h2><a href="#{tag.id}">{tag.name}</a></h2>
        {#if tag.description}<div class="prose">{@html tag.description}</div>{/if}
      </div>
      {#each tag.operations as op (op.id)}
        <ApiOperation {op} />
      {/each}
    </section>
  {/each}

  {#if ref.models.length}
    <section class="tag models" id="models">
      <div class="tag-head"><h2><a href="#models">Models</a></h2></div>
      <div class="model-list">
        {#each ref.models as m (m.id)}
          <ApiModel model={m} />
        {/each}
      </div>
    </section>
  {/if}
</div>

<ApiClient />

<style>
  .ref { min-width: 0; }
  .intro {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: 3rem;
    padding-bottom: 3rem;
  }
  .badges { display: flex; gap: 0.4rem; margin-bottom: 0.75rem; }
  .badge {
    padding: 0.1rem 0.5rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--muted-fg);
  }
  h1 { margin: 0 0 1rem; font-size: 2.1rem; font-weight: 600; letter-spacing: -0.025em; line-height: 1.15; }
  .intro-side { display: flex; flex-direction: column; gap: 0.75rem; }
  .card {
    padding: 0.85rem 1rem;
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    background: var(--card);
  }
  .card-title {
    margin: 0 0 0.5rem;
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted-fg);
  }
  .url { font-family: var(--font-mono); font-size: 0.82rem; overflow-wrap: anywhere; }
  .hint { margin: 0.4rem 0 0; font-size: 0.75rem; color: var(--muted-fg); }
  .hint :global(p) { margin: 0.2rem 0; }
  select, input {
    width: 100%;
    padding: 0.4rem 0.6rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--fg);
    font: inherit;
    font-size: 0.82rem;
  }
  .field { display: flex; flex-direction: column; gap: 0.25rem; margin-top: 0.5rem; font-size: 0.75rem; color: var(--muted-fg); }
  .field span { text-transform: capitalize; }
  .clients { display: flex; flex-wrap: wrap; gap: 0.3rem; }
  .clients button {
    padding: 0.3rem 0.65rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    font-size: 0.78rem;
    color: var(--muted-fg);
    cursor: pointer;
  }
  .clients button:hover { color: var(--fg); }
  .clients button[aria-checked='true'] { color: var(--fg); border-color: var(--fg); }
  .tag { padding-top: 1rem; }
  .tag-head { padding: 2rem 0 1.5rem; border-top: 1px solid var(--border); max-width: calc(50% - 1.5rem); }
  .tag-head h2 { margin: 0 0 0.5rem; font-size: 1.6rem; font-weight: 600; letter-spacing: -0.02em; }
  .tag-head h2 a { text-decoration: none; }
  .model-list { display: flex; flex-direction: column; gap: 0.5rem; padding-bottom: 3rem; }
  @media (max-width: 1100px) {
    .intro { grid-template-columns: minmax(0, 1fr); gap: 1.5rem; }
    .tag-head { max-width: none; }
  }
</style>
