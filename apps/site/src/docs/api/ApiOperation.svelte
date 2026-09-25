<script lang="ts">
  import {
    applyAuth,
    buildRequest,
    clients,
    defaultInput,
    prettyJSON,
    snippet,
    statusTone,
    type ApiOperation,
  } from '../../lib/openapi'
  import CodePanel from './CodePanel.svelte'
  import MethodBadge from './MethodBadge.svelte'
  import SchemaProp from './SchemaProp.svelte'
  import SchemaView from './SchemaView.svelte'
  import { getApi } from './state.svelte'

  let { op }: { op: ApiOperation } = $props()
  const api = getApi()

  const groups = $derived(
    (
      [
        ['path', 'Path Parameters'],
        ['query', 'Query Parameters'],
        ['header', 'Headers'],
        ['cookie', 'Cookies'],
      ] as const
    )
      .map(([where, title]) => ({ title, params: op.parameters.filter((p) => p.in === where) }))
      .filter((g) => g.params.length),
  )

  let bodyType = $state(0)
  let bodyExample = $state(0)
  const bodyMedia = $derived(op.request_body?.content?.[bodyType] ?? op.request_body?.content?.[0])

  // Request sample: the reader's language, or one of the spec's own samples.
  let custom = $state(-1)
  const request = $derived.by(() => {
    const input = defaultInput(op, api.server)
    if (bodyMedia) {
      const ex = bodyMedia.examples?.[bodyExample]?.value ?? bodyMedia.examples?.[0]?.value
      input.contentType = bodyMedia.type
      input.body = ex === undefined ? '' : typeof ex === 'string' && !bodyMedia.type.includes('json') ? ex : prettyJSON(ex)
    }
    return buildRequest(op, input, applyAuth(op, api.ref.security_schemes, api.auth, true))
  })
  const sampleLang = $derived(custom >= 0 ? (op.code_samples?.[custom]?.lang ?? '') : (clients.find((c) => c.id === api.client)?.lang ?? 'shell'))
  const sample = $derived(custom >= 0 ? (op.code_samples?.[custom]?.source ?? '') : snippet(api.client, request))
  function pick(e: Event) {
    const v = (e.currentTarget as HTMLSelectElement).value
    if (v.startsWith('custom:')) custom = Number(v.slice(7))
    else {
      custom = -1
      api.setClient(v)
    }
  }

  // Response sample: a tab per status code, then its examples.
  let status = $state(0)
  let exampleIndex = $state(0)
  const response = $derived(op.responses[status] ?? op.responses[0])
  const responseMedia = $derived(response?.content?.[0])
  const responseExample = $derived(responseMedia?.examples?.[exampleIndex] ?? responseMedia?.examples?.[0])

  const security = $derived(op.security.map((k) => api.ref.security_schemes.find((s) => s.key === k)).filter((s) => !!s))
</script>

<section class="op" id={op.id} data-op>
  <div class="left">
    <h3 class="title">
      <a href="#{op.id}" class="anchor">{op.summary}</a>
      {#if op.deprecated}<span class="deprecated">Deprecated</span>{/if}
    </h3>
    <div class="endpoint">
      <MethodBadge method={op.method} />
      <code>{op.path}</code>
    </div>
    {#if op.description}<div class="prose desc">{@html op.description}</div>{/if}

    {#if security.length}
      <div class="auth-note">
        Requires {security.map((s) => s.key).join(' or ')}
      </div>
    {/if}

    {#each groups as g (g.title)}
      <div class="block">
        <h4>{g.title}</h4>
        <ul class="props">
          {#each g.params as p (p.name)}
            <SchemaProp
              name={p.name}
              schema={p.schema}
              required={p.required}
              description={p.description}
              deprecated={p.deprecated}
              example={p.example} />
          {/each}
        </ul>
      </div>
    {/each}

    {#if op.request_body}
      <div class="block">
        <h4>
          Body
          {#if op.request_body.required}<span class="required">required</span>{/if}
          <span class="spacer"></span>
          {#if op.request_body.content.length > 1}
            <select class="mini" bind:value={bodyType} onchange={() => (bodyExample = 0)} aria-label="Content type">
              {#each op.request_body.content as m, i (m.type)}<option value={i}>{m.type}</option>{/each}
            </select>
          {:else if bodyMedia}
            <span class="ctype">{bodyMedia.type}</span>
          {/if}
        </h4>
        {#if op.request_body.description}<div class="prose desc">{@html op.request_body.description}</div>{/if}
        {#if bodyMedia?.schema}<SchemaView schema={bodyMedia.schema} />{/if}
      </div>
    {/if}

    {#if op.responses.length}
      <div class="block">
        <h4>Responses</h4>
        <div class="responses">
          {#each op.responses as r (r.status)}
            {@const media = r.content?.[0]}
            <details class="response">
              <summary>
                <span class="status" data-tone={statusTone(r.status)}>{r.status}</span>
                <span class="rdesc">{@html r.description ?? ''}</span>
                {#if media}<span class="ctype">{media.type}</span>{/if}
              </summary>
              <div class="rbody">
                {#if r.headers?.length}
                  <p class="sub">Headers</p>
                  <ul class="props">
                    {#each r.headers as h (h.name)}
                      <SchemaProp name={h.name} schema={h.schema} description={h.description} />
                    {/each}
                  </ul>
                {/if}
                {#if media?.schema}
                  <SchemaView schema={media.schema} />
                {:else if !r.headers?.length}
                  <p class="sub">No body</p>
                {/if}
              </div>
            </details>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <div class="right">
    <div class="sticky">
      <CodePanel code={sample} lang={sampleLang}>
        {#snippet head()}
          <MethodBadge method={op.method} />
          <code class="hpath">{op.path}</code>
          <select class="mini" value={custom >= 0 ? `custom:${custom}` : api.client} onchange={pick} aria-label="Client">
            {#each op.code_samples ?? [] as cs, i (i)}
              <option value="custom:{i}">{cs.label || cs.lang}</option>
            {/each}
            {#each clients as c (c.id)}<option value={c.id}>{c.label}</option>{/each}
          </select>
        {/snippet}
        {#snippet foot()}
          {#if (bodyMedia?.examples?.length ?? 0) > 1}
            <select class="mini example" bind:value={bodyExample} aria-label="Request example">
              {#each bodyMedia?.examples ?? [] as ex, i (i)}<option value={i}>{ex.summary || ex.name}</option>{/each}
            </select>
          {/if}
          <button class="test" type="button" onclick={() => (api.testing = op)}>
            <svg viewBox="0 0 24 24" width="12" height="12" fill="currentColor" aria-hidden="true"><path d="M7 4v16l13-8z" /></svg>
            Test Request
          </button>
        {/snippet}
      </CodePanel>

      {#if op.responses.length}
        <CodePanel code={prettyJSON(responseExample?.value)} lang={responseMedia?.type ?? ''}>
          {#snippet head()}
            <div class="tabs" role="tablist" aria-label="Response status">
              {#each op.responses as r, i (r.status)}
                <button
                  type="button"
                  role="tab"
                  aria-selected={i === status}
                  data-tone={statusTone(r.status)}
                  onclick={() => {
                    status = i
                    exampleIndex = 0
                  }}>{r.status}</button>
              {/each}
            </div>
            {#if (responseMedia?.examples?.length ?? 0) > 1}
              <select class="mini" bind:value={exampleIndex} aria-label="Example">
                {#each responseMedia?.examples ?? [] as ex, i (i)}<option value={i}>{ex.summary || ex.name}</option>{/each}
              </select>
            {/if}
          {/snippet}
        </CodePanel>
      {/if}
    </div>
  </div>
</section>

<style>
  .op {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: 3rem;
    padding: 3rem 0;
    border-top: 1px solid var(--border);
  }
  .title {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.6rem;
    margin: 0 0 0.6rem;
    font-size: 1.25rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    line-height: 1.3;
  }
  .anchor { text-decoration: none; }
  .anchor:hover { text-decoration: underline; text-underline-offset: 4px; text-decoration-color: var(--border); }
  .deprecated {
    font-size: 0.7rem;
    font-weight: 500;
    padding: 0.1rem 0.45rem;
    border-radius: 999px;
    color: var(--warn);
    background: color-mix(in oklab, var(--warn) 14%, transparent);
  }
  .endpoint { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem; }
  .endpoint code { font-family: var(--font-mono); font-size: 0.82rem; color: var(--muted-fg); overflow-wrap: anywhere; }
  .desc { font-size: 0.925rem; }
  .desc :global(p:first-child) { margin-top: 0; }
  .auth-note {
    display: inline-block;
    margin: 0.25rem 0 0.5rem;
    padding: 0.2rem 0.6rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    font-size: 0.75rem;
    color: var(--muted-fg);
  }
  .block { margin-top: 1.75rem; }
  h4 {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0;
    padding-bottom: 0.6rem;
    border-bottom: 1px solid var(--border);
    font-size: 0.95rem;
    font-weight: 600;
  }
  .spacer { flex: 1; }
  .required { font-size: 0.72rem; font-weight: 500; color: var(--warn); }
  .ctype { font-family: var(--font-mono); font-size: 0.72rem; font-weight: 400; color: var(--muted-fg); }
  .props { list-style: none; margin: 0; padding: 0; }
  .mini {
    max-width: 11rem;
    padding: 0.15rem 0.4rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--fg);
    font: inherit;
    font-size: 0.75rem;
    font-weight: 400;
  }
  .responses { border: 1px solid var(--border); border-radius: 0.6rem; margin-top: 0.75rem; overflow: hidden; }
  .response + .response { border-top: 1px solid var(--border); }
  .response summary {
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
    padding: 0.6rem 0.85rem;
    cursor: pointer;
    list-style: none;
    font-size: 0.85rem;
  }
  .response summary::-webkit-details-marker { display: none; }
  .response summary::before {
    content: '+';
    font-family: var(--font-mono);
    color: var(--muted-fg);
    width: 0.7rem;
  }
  .response[open] summary::before { content: '−'; }
  .response summary:hover { background: var(--muted); }
  .rdesc { flex: 1; color: var(--muted-fg); }
  .rdesc :global(p) { margin: 0; display: inline; }
  .rbody { padding: 0 0.85rem 0.5rem; border-top: 1px solid var(--border-soft); }
  .sub { margin: 0.6rem 0 0.2rem; font-size: 0.75rem; color: var(--muted-fg); }
  .status { font-family: var(--font-mono); font-weight: 600; font-size: 0.8rem; }
  [data-tone='ok'] { color: var(--success); }
  [data-tone='redirect'] { color: var(--info); }
  [data-tone='client'] { color: var(--warn); }
  [data-tone='server'] { color: var(--error); }

  .right { min-width: 0; }
  .sticky {
    position: sticky;
    top: calc(var(--nav-h) + 1.5rem);
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .hpath {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: 0.78rem;
    color: var(--muted-fg);
  }
  .tabs { display: flex; flex-wrap: wrap; gap: 0.15rem; flex: 1; }
  .tabs button {
    padding: 0.2rem 0.5rem;
    border: 0;
    border-radius: 6px;
    background: none;
    font-family: var(--font-mono);
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    opacity: 0.65;
  }
  .tabs button[aria-selected='true'] { background: var(--card); opacity: 1; box-shadow: 0 0 0 1px var(--border); }
  .test {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    height: 1.9rem;
    padding: 0 0.75rem;
    border: 0;
    border-radius: 6px;
    background: var(--primary);
    color: var(--primary-fg);
    font-size: 0.78rem;
    font-weight: 500;
    cursor: pointer;
  }
  .test:hover { opacity: 0.9; }
  .example { margin-right: auto; }

  @media (max-width: 1100px) {
    .op { grid-template-columns: minmax(0, 1fr); gap: 1.5rem; padding: 2.25rem 0; }
    .sticky { position: static; }
  }
</style>
