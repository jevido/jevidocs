<script lang="ts">
  import Icon from '../../lib/Icon.svelte'
  import {
    applyAuth,
    buildRequest,
    defaultInput,
    highlight,
    statusTone,
    type ApiOperation,
    type ApiParameter,
  } from '../../lib/openapi'
  import MethodBadge from './MethodBadge.svelte'
  import { getApi } from './state.svelte'

  // The request client behind "Test Request": edit parameters, headers and
  // the body of one operation, send it from the browser and read the answer.
  const api = getApi()
  let dialog = $state<HTMLDialogElement>()

  type Row = { name: string; value: string; on: boolean; param?: ApiParameter }
  let op = $state.raw<ApiOperation | null>(null)
  let path = $state<Row[]>([])
  let query = $state<Row[]>([])
  let headers = $state<Row[]>([])
  let cookies = $state<Row[]>([])
  let body = $state('')
  let tab = $state<'params' | 'headers' | 'body'>('params')

  type Result = { status: number; statusText: string; ms: number; size: number; headers: [string, string][]; body: string; type: string }
  let result = $state.raw<Result | null>(null)
  let error = $state('')
  let sending = $state(false)
  let responseTab = $state<'body' | 'headers'>('body')
  let ctrl: AbortController | undefined

  // Opening: api.testing is set by an operation's Test Request button.
  $effect(() => {
    const next = api.testing
    if (!next || !dialog) return
    load(next)
    if (!dialog.open) dialog.showModal()
  })

  function load(o: ApiOperation) {
    op = o
    const d = defaultInput(o, api.server, false)
    const val = (p: ApiParameter) => {
      const ex = p.example ?? p.schema?.example ?? p.schema?.default
      return ex === undefined || ex === null ? '' : typeof ex === 'string' ? ex : JSON.stringify(ex)
    }
    const rows = (where: string) =>
      o.parameters.filter((p) => p.in === where).map((p) => ({ name: p.name, value: val(p), on: !!p.required || val(p) !== '', param: p }))
    path = rows('path')
    query = rows('query')
    headers = rows('header')
    cookies = rows('cookie')
    body = d.body ?? ''
    tab = o.parameters.length || !o.request_body ? 'params' : 'body'
    result = null
    error = ''
  }

  function close() {
    ctrl?.abort()
    dialog?.close()
  }

  const pairs = (rows: Row[]) => rows.filter((r) => r.on && r.name).map((r) => [r.name, r.value] as [string, string])

  const request = $derived.by(() => {
    if (!op) return null
    const pathVals: Record<string, string> = {}
    for (const r of path) pathVals[r.name] = r.value
    const input = {
      server: api.server,
      path: pathVals,
      query: pairs(query),
      headers: pairs(headers),
      cookies: pairs(cookies),
      body: op.request_body ? body : undefined,
      contentType: op.request_body?.content?.[0]?.type,
    }
    return buildRequest(op, input, applyAuth(op, api.ref.security_schemes, api.auth, false))
  })

  async function send() {
    if (!request || sending) return
    ctrl?.abort()
    ctrl = new AbortController()
    sending = true
    error = ''
    const started = performance.now()
    try {
      // Cookies cannot be set from a page; browsers send their own.
      const hdrs = request.headers.filter(([k]) => k.toLowerCase() !== 'cookie')
      const res = await fetch(request.url, { method: request.method, headers: hdrs, body: request.body, signal: ctrl.signal })
      const text = await res.text()
      const type = res.headers.get('content-type') ?? ''
      let pretty = text
      if (type.includes('json')) {
        try {
          pretty = JSON.stringify(JSON.parse(text), null, 2)
        } catch {
          // not valid JSON after all
        }
      }
      result = {
        status: res.status,
        statusText: res.statusText,
        ms: Math.round(performance.now() - started),
        size: new Blob([text]).size,
        headers: [...res.headers.entries()],
        body: pretty,
        type,
      }
      responseTab = 'body'
    } catch (e) {
      if ((e as Error).name === 'AbortError') return
      result = null
      error = `The request failed: ${(e as Error).message}. The server may be unreachable, or it does not allow requests from this site (CORS).`
    } finally {
      sending = false
    }
  }

  function onkeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      void send()
    }
  }

  const size = (n: number) => (n < 1024 ? `${n} B` : `${(n / 1024).toFixed(1)} KB`)
  const lang = (type: string) => (type.includes('json') ? 'json' : '')
</script>

{#snippet rowsTable(title: string, rows: Row[], fixed: boolean)}
  <div class="group">
    <p class="group-title">{title}</p>
    <div class="rows">
      {#each rows as r, i (i)}
        <div class="row">
          <input type="checkbox" bind:checked={r.on} aria-label="Send {r.name || 'this row'}" disabled={fixed} />
          {#if r.param}
            <span class="key" title={r.param.description ? r.param.description.replace(/<[^>]+>/g, '') : undefined}>
              {r.name}{#if r.param.required}<span class="req">*</span>{/if}
            </span>
          {:else}
            <input class="key-input" bind:value={r.name} placeholder="Name" />
          {/if}
          {#if r.param?.schema?.enum?.length}
            <select bind:value={r.value} onchange={() => (r.on = true)}>
              <option value=""></option>
              {#each r.param.schema.enum as e (String(e))}<option value={String(e)}>{String(e)}</option>{/each}
            </select>
          {:else}
            <input bind:value={r.value} placeholder={r.param?.schema?.type ?? 'Value'} oninput={() => (r.on = true)} />
          {/if}
        </div>
      {/each}
      {#if !fixed}
        <button class="add" type="button" onclick={() => rows.push({ name: '', value: '', on: true })}>+ Add</button>
      {/if}
    </div>
  </div>
{/snippet}

<dialog bind:this={dialog} class="client" aria-label="Test request" onclose={() => (api.testing = null)} {onkeydown}>
  {#if op && request}
    <header class="bar">
      <MethodBadge method={op.method} />
      <code class="address" title={request.url}>{request.url}</code>
      <button class="send" type="button" onclick={send} disabled={sending}>
        {sending ? 'Sending…' : 'Send'}
        <kbd>⌘↵</kbd>
      </button>
      <button class="icon-btn" type="button" aria-label="Close" onclick={close}><Icon name="x" size={16} /></button>
    </header>
    <div class="split">
      <div class="pane">
        <p class="op-title">{op.summary}</p>
        <div class="tabs" role="tablist">
          <button type="button" role="tab" aria-selected={tab === 'params'} onclick={() => (tab = 'params')}>Parameters</button>
          <button type="button" role="tab" aria-selected={tab === 'headers'} onclick={() => (tab = 'headers')}>Headers</button>
          {#if op.request_body}
            <button type="button" role="tab" aria-selected={tab === 'body'} onclick={() => (tab = 'body')}>Body</button>
          {/if}
        </div>
        {#if tab === 'params'}
          {#if path.length}{@render rowsTable('Path', path, true)}{/if}
          {@render rowsTable('Query', query, false)}
          {#if cookies.length}{@render rowsTable('Cookies', cookies, false)}{/if}
        {:else if tab === 'headers'}
          {@render rowsTable('Headers', headers, false)}
          {#if request.headers.some(([k]) => k === 'Authorization') || op.security.length}
            <p class="hint">Credentials from the Authentication panel are added for you.</p>
          {/if}
        {:else}
          <div class="group">
            <p class="group-title">{op.request_body?.content?.[0]?.type ?? 'Body'}</p>
            <textarea bind:value={body} spellcheck="false" rows="14"></textarea>
          </div>
        {/if}
      </div>
      <div class="pane response">
        {#if error}
          <p class="error">{error}</p>
        {:else if result}
          <div class="meta">
            <span class="status" data-tone={statusTone(result.status)}>{result.status} {result.statusText}</span>
            <span>{result.ms} ms</span>
            <span>{size(result.size)}</span>
          </div>
          <div class="tabs" role="tablist">
            <button type="button" role="tab" aria-selected={responseTab === 'body'} onclick={() => (responseTab = 'body')}>Body</button>
            <button type="button" role="tab" aria-selected={responseTab === 'headers'} onclick={() => (responseTab = 'headers')}>Headers ({result.headers.length})</button>
          </div>
          {#if responseTab === 'body'}
            <pre class="chroma out"><code>{@html highlight(result.body, lang(result.type))}</code></pre>
          {:else}
            <dl class="hdrs">
              {#each result.headers as [k, v] (k)}<dt>{k}</dt><dd>{v}</dd>{/each}
            </dl>
          {/if}
        {:else}
          <p class="empty">Send the request to see the response.</p>
        {/if}
      </div>
    </div>
  {/if}
</dialog>

<style>
  .client {
    width: min(1100px, calc(100vw - 2rem));
    height: min(720px, calc(100vh - 2rem));
    padding: 0;
    border: 1px solid var(--border);
    border-radius: 0.9rem;
    background: var(--bg);
    color: var(--fg);
    box-shadow: var(--shadow);
    overflow: hidden;
  }
  .client[open] { display: flex; flex-direction: column; }
  .client:focus-visible { outline: none; }
  .client::backdrop { background: var(--overlay); }
  .bar {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.6rem 0.75rem;
    border-bottom: 1px solid var(--border);
  }
  .address {
    flex: 1;
    min-width: 0;
    padding: 0.4rem 0.65rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--card);
    font-family: var(--font-mono);
    font-size: 0.8rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .send {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    height: 2rem;
    padding: 0 0.85rem;
    border: 0;
    border-radius: 6px;
    background: var(--primary);
    color: var(--primary-fg);
    font-size: 0.82rem;
    font-weight: 500;
    cursor: pointer;
  }
  .send:disabled { opacity: 0.6; cursor: progress; }
  .send kbd { background: transparent; color: inherit; border-color: color-mix(in oklab, var(--primary-fg) 40%, transparent); }
  .split { flex: 1; min-height: 0; display: grid; grid-template-columns: 1fr 1fr; }
  .pane { min-width: 0; overflow-y: auto; padding: 0.9rem 1rem; scrollbar-width: thin; }
  .response { border-left: 1px solid var(--border); background: var(--card); }
  .op-title { margin: 0 0 0.6rem; font-weight: 600; }
  .tabs { display: flex; gap: 0.2rem; margin-bottom: 0.75rem; border-bottom: 1px solid var(--border); }
  .tabs button {
    padding: 0.4rem 0.6rem;
    border: 0;
    border-bottom: 2px solid transparent;
    background: none;
    font-size: 0.8rem;
    color: var(--muted-fg);
    cursor: pointer;
    margin-bottom: -1px;
  }
  .tabs button[aria-selected='true'] { color: var(--fg); border-bottom-color: var(--fg); }
  .group { margin-bottom: 1rem; }
  .group-title { margin: 0 0 0.4rem; font-size: 0.72rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--muted-fg); }
  .rows { border: 1px solid var(--border); border-radius: 6px; overflow: hidden; }
  .row { display: grid; grid-template-columns: auto minmax(6rem, 0.8fr) 1.2fr; align-items: center; gap: 0.5rem; padding: 0.25rem 0.5rem; border-bottom: 1px solid var(--border-soft); }
  .key { font-family: var(--font-mono); font-size: 0.78rem; overflow-wrap: anywhere; }
  .req { color: var(--warn); margin-left: 0.1rem; }
  .row input:not([type='checkbox']), .row select, textarea {
    width: 100%;
    padding: 0.3rem 0.45rem;
    border: 1px solid transparent;
    border-radius: 4px;
    background: transparent;
    color: var(--fg);
    font-family: var(--font-mono);
    font-size: 0.78rem;
  }
  .row input:not([type='checkbox']):focus, .row select:focus { border-color: var(--border); background: var(--bg); outline: none; }
  textarea { border-color: var(--border); background: var(--card); line-height: 1.55; resize: vertical; }
  .add { width: 100%; padding: 0.35rem; border: 0; background: none; font-size: 0.78rem; color: var(--muted-fg); cursor: pointer; text-align: left; }
  .add:hover { color: var(--fg); background: var(--muted); }
  .hint { font-size: 0.75rem; color: var(--muted-fg); }
  .meta { display: flex; gap: 1rem; margin-bottom: 0.6rem; font-size: 0.8rem; color: var(--muted-fg); font-family: var(--font-mono); }
  .status { font-weight: 600; }
  [data-tone='ok'] { color: var(--success); }
  [data-tone='redirect'] { color: var(--info); }
  [data-tone='client'] { color: var(--warn); }
  [data-tone='server'] { color: var(--error); }
  .out { margin: 0; font-family: var(--font-mono); font-size: 0.78rem; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere; }
  .hdrs { display: grid; grid-template-columns: auto 1fr; gap: 0.3rem 1rem; margin: 0; font-family: var(--font-mono); font-size: 0.75rem; }
  .hdrs dt { color: var(--muted-fg); }
  .hdrs dd { margin: 0; overflow-wrap: anywhere; }
  .empty { margin-top: 3rem; text-align: center; font-size: 0.85rem; color: var(--muted-fg); }
  .error { font-size: 0.85rem; color: var(--error); }
  @media (max-width: 800px) {
    .client { width: 100vw; height: 100vh; max-width: none; max-height: none; border-radius: 0; }
    .split { grid-template-columns: 1fr; grid-template-rows: auto auto; overflow-y: auto; }
    .pane { overflow: visible; }
    .response { border-left: 0; border-top: 1px solid var(--border); }
  }
</style>
