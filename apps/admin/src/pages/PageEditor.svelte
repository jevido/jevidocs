<script lang="ts">
  import { api } from '../lib/api.svelte'
  import { go, href, router } from '../lib/router.svelte'
  import { startTranslation, takeTranslation } from '../lib/translation'
  import { toast } from '../lib/toast.svelte'
  import { formatDate, slugify } from '../lib/format'
  import { snippets } from '../lib/snippets'
  import Preview from '../lib/Preview.svelte'
  import HistoryDrawer from '../lib/HistoryDrawer.svelte'
  import { filesFrom, uploadInto } from '../lib/paste-upload'
  import type { AdminPage, AdminPageInput, TocItem } from '../lib/types'
  import { untrack } from 'svelte'

  let { project, id }: { project: string; id: number | 'new' } = $props()

  // App.svelte keys this component on project/id, so they never change here.
  const initialId = untrack(() => id)
  const initialProject = untrack(() => project)

  const empty: AdminPageInput = {
    slug: '',
    title: '',
    description: '',
    icon: '',
    position: 0,
    section: '',
    published: true,
    body: '# Title\n\nStart writing…\n',
    locale: '',
  }

  const draft = initialId === 'new' ? takeTranslation() : null
  let form = $state<AdminPageInput>(draft ?? { ...empty })
  let saved = $state(JSON.stringify(draft ? { ...empty } : empty))

  // Other languages of the project this page could be translated into.
  let locales = $state<string[]>([])
  let defaultLocale = $state('')
  api
    .projects()
    .then((all) => {
      const p = all.find((x) => x.slug === initialProject)
      locales = p?.locales ?? []
      defaultLocale = p?.default_locale ?? ''
    })
    .catch(() => {})
  const missingLocales = $derived(
    locales.filter((l) => l !== (form.locale || defaultLocale)),
  )

  function translate(locale: string) {
    startTranslation({ ...form, locale: locale === defaultLocale ? '' : locale })
    go(`/projects/${initialProject}/pages/new`)
  }
  let loading = $state(initialId !== 'new')
  let busy = $state(false)
  let updatedAt = $state('')
  let html = $state('')
  let toc = $state.raw<TocItem[]>([])
  let previewError = $state('')
  let view = $state<'split' | 'write' | 'preview'>('split')
  let textarea = $state<HTMLTextAreaElement>()
  let slugTouched = $state(initialId !== 'new')
  let historyOpen = $state(false)

  function onRestored(p: AdminPage) {
    const input: AdminPageInput = {
      slug: p.slug,
      title: p.title,
      description: p.description,
      icon: p.icon,
      position: p.position,
      section: p.section,
      published: p.published,
      body: p.body,
      locale: p.locale ?? '',
    }
    form = input
    saved = JSON.stringify(input)
    updatedAt = p.updated_at
    historyOpen = false
  }

  const dirty = $derived(JSON.stringify(form) !== saved)
  const isNew = $derived(id === 'new')

  if (initialId !== 'new') {
    api
      .page(initialProject, initialId)
      .then((p) => {
        const input: AdminPageInput = {
          slug: p.slug,
          title: p.title,
          description: p.description,
          icon: p.icon,
          position: p.position,
          section: p.section,
          published: p.published,
          body: p.body,
          locale: p.locale ?? '',
        }
        form = input
        saved = JSON.stringify(input)
        updatedAt = p.updated_at
      })
      .catch(toast.error)
      .finally(() => (loading = false))
  }

  router.guard = () => !dirty || confirm('You have unsaved changes. Leave anyway?')

  // Debounced preview of the body, rendered by the API.
  $effect(() => {
    const body = form.body
    const timer = setTimeout(() => {
      api
        .preview(body)
        .then((r) => {
          html = r.html
          toc = r.toc ?? []
          previewError = ''
        })
        .catch((e) => (previewError = e.message))
    }, 400)
    return () => clearTimeout(timer)
  })

  function onTitle() {
    if (!slugTouched && isNew) form.slug = slugify(form.title)
  }

  async function save() {
    if (busy) return
    if (!form.title.trim()) {
      toast.error('A title is required')
      return
    }
    busy = true
    const input = $state.snapshot(form)
    input.position = Number(input.position) || 0
    try {
      if (id === 'new') {
        const p = await api.createPage(project, input)
        saved = JSON.stringify(form)
        toast.ok('Page created')
        router.guard = null
        go(`/projects/${project}/pages/${p.id}`)
      } else {
        const p = await api.updatePage(project, id, input)
        saved = JSON.stringify(form)
        updatedAt = p.updated_at
        toast.ok('Saved')
      }
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  async function remove() {
    if (id === 'new') return
    if (!confirm(`Delete "${form.title}"?`)) return
    try {
      await api.deletePage(project, id)
      router.guard = null
      toast.ok('Page deleted')
      go(`/projects/${project}`)
    } catch (err) {
      toast.error(err)
    }
  }

  function insert(text: string) {
    const el = textarea
    if (!el) return
    const start = el.selectionStart
    const before = el.value.slice(0, start)
    const prefix = before.length && !before.endsWith('\n\n') ? (before.endsWith('\n') ? '\n' : '\n\n') : ''
    el.focus()
    el.setRangeText(prefix + text, start, el.selectionEnd, 'end')
    form.body = el.value
  }

  function onFiles(e: ClipboardEvent | DragEvent) {
    const files = filesFrom(e)
    if (!files.length || !textarea) return
    e.preventDefault()
    if (e instanceof DragEvent) textarea.focus()
    uploadInto(textarea, project, files, (v) => (form.body = v))
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Tab' && !e.shiftKey && !e.ctrlKey && !e.metaKey && !e.altKey) {
      e.preventDefault()
      const el = e.currentTarget as HTMLTextAreaElement
      el.setRangeText('  ', el.selectionStart, el.selectionEnd, 'end')
      form.body = el.value
    }
  }

  function onWindowKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
      e.preventDefault()
      save()
    }
  }

  function onBeforeUnload(e: BeforeUnloadEvent) {
    if (dirty) e.preventDefault()
  }
</script>

<svelte:window onkeydown={onWindowKeydown} onbeforeunload={onBeforeUnload} />

{#if historyOpen && id !== 'new'}
  <HistoryDrawer {project} pageId={id} current={form.body} {dirty} onclose={() => (historyOpen = false)} onrestore={onRestored} />
{/if}

<div class="crumbs muted">
  <a href={href('/projects')}>Projects</a> /
  <a href={href(`/projects/${project}`)}>{project}</a> /
  {isNew ? 'New page' : form.slug || 'index'}
</div>

<div class="page-head">
  <div>
    <h1>{form.title || (isNew ? 'New page' : 'Untitled')}</h1>
    <p>
      {#if dirty}<span class="badge warn">Unsaved changes</span>{:else if !isNew}<span class="badge ok">Saved</span>{/if}
      {#if updatedAt}<span class="muted"> · updated {formatDate(updatedAt)}</span>{/if}
    </p>
  </div>
  <div class="spacer"></div>
  {#if !isNew && locales.length > 1}
    {#each missingLocales as l (l)}
      <button class="btn" onclick={() => translate(l)} title="Start a {l} translation of this page">Translate: {l}</button>
    {/each}
  {/if}
  {#if !isNew}<button class="btn" onclick={() => (historyOpen = true)}>History</button>{/if}
  {#if !isNew}<button class="btn danger" onclick={remove}>Delete</button>{/if}
  <button class="btn primary" onclick={save} disabled={busy || loading || (!dirty && !isNew)}>
    {busy ? 'Saving…' : isNew ? 'Create page' : 'Save'} <kbd>⌘S</kbd>
  </button>
</div>

{#if loading}
  <p class="muted">Loading…</p>
{:else}
  <details class="card meta" open={isNew}>
    <summary>Page settings <span class="muted">/{form.slug}</span></summary>
    <div class="grid">
      <label class="field">
        Title
        <input type="text" required bind:value={form.title} oninput={onTitle} />
      </label>
      <label class="field">
        Slug
        <input type="text" placeholder="guides/install" bind:value={form.slug} oninput={() => (slugTouched = true)} />
        <span class="hint">Path in the tree. Empty = project index. A folder's index uses the folder path.</span>
      </label>
      <label class="field wide">
        Description
        <input type="text" bind:value={form.description} />
      </label>
      <label class="field">
        Icon
        <input type="text" placeholder="book, rocket, …" bind:value={form.icon} />
      </label>
      <label class="field">
        Position
        <input type="number" bind:value={form.position} />
        <span class="hint">Lower comes first among siblings.</span>
      </label>
      <label class="field">
        Locale
        <input type="text" placeholder="default" bind:value={form.locale} disabled={!isNew} />
        <span class="hint">Empty = default language. A translation is a new page with the same slug and a locale.</span>
      </label>
      <label class="field">
        Section
        <input type="text" placeholder="Guides" bind:value={form.section} />
        <span class="hint">Top-level pages: starts a sidebar separator.</span>
      </label>
      <label class="check">
        <input type="checkbox" bind:checked={form.published} />
        Published
      </label>
    </div>
  </details>

  <div class="editor card">
    <div class="bar">
      <div class="snips">
        {#each snippets as s (s.label)}
          <button class="btn sm ghost" type="button" onclick={() => insert(s.text)}>{s.label}</button>
        {/each}
      </div>
      <div class="spacer"></div>
      <div class="seg" role="group" aria-label="View">
        {#each ['write', 'split', 'preview'] as const as v (v)}
          <button type="button" aria-pressed={view === v} onclick={() => (view = v)}>{v}</button>
        {/each}
      </div>
    </div>
    <div class={['panes', view]}>
      {#if view !== 'preview'}
        <textarea
          bind:this={textarea}
          bind:value={form.body}
          onkeydown={onKeydown}
          onpaste={onFiles}
          ondrop={onFiles}
          spellcheck="false"
          aria-label="Markdown"
        ></textarea>
      {/if}
      {#if view !== 'write'}
        <div class="preview">
          {#if previewError}<p class="err">Preview failed: {previewError}</p>{/if}
          <Preview {html} />
          {#if toc.length}
            <div class="toc muted">
              <strong>On this page</strong>
              {#each toc as t (t.url)}
                <div style:padding-left="{(t.depth - 2) * 0.75}rem">{t.title}</div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .crumbs {
    font-size: 0.85rem;
    margin-bottom: 0.5rem;
  }
  .crumbs a {
    text-decoration: none;
  }
  .page-head p {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }
  kbd {
    font-size: 0.7rem;
    opacity: 0.6;
  }
  .meta {
    margin-bottom: 1rem;
  }
  .meta summary {
    padding: 0.8rem 1.1rem;
    cursor: pointer;
    font-weight: 500;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
    gap: 1rem;
    padding: 0 1.1rem 1.1rem;
  }
  .wide {
    grid-column: 1 / -1;
  }
  .editor {
    overflow: hidden;
  }
  .bar {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.5rem;
    border-bottom: 1px solid var(--border);
    background: var(--surface-2);
    flex-wrap: wrap;
  }
  .snips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.1rem;
  }
  .seg {
    display: flex;
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }
  .seg button {
    font: inherit;
    font-size: 0.8rem;
    text-transform: capitalize;
    border: none;
    background: var(--surface);
    color: var(--muted);
    padding: 0.25rem 0.65rem;
    cursor: pointer;
  }
  .seg button[aria-pressed='true'] {
    background: var(--primary);
    color: var(--primary-text);
  }
  .panes {
    display: grid;
    height: calc(100vh - 16rem);
    min-height: 28rem;
  }
  .panes.split {
    grid-template-columns: 1fr 1fr;
  }
  textarea {
    border: none;
    border-radius: 0;
    resize: none;
    height: 100%;
    padding: 1rem 1.1rem;
    font-family: var(--mono);
    font-size: 13px;
    line-height: 1.65;
    tab-size: 2;
    background: var(--surface);
  }
  textarea:focus-visible {
    outline: none;
  }
  .split textarea {
    border-right: 1px solid var(--border);
  }
  .preview {
    overflow: auto;
    padding: 1.25rem 1.5rem;
    min-width: 0;
  }
  .err {
    color: var(--danger);
    font-size: 0.85rem;
  }
  .toc {
    margin-top: 2rem;
    padding-top: 1rem;
    border-top: 1px dashed var(--border);
    font-size: 0.8rem;
  }
  @media (max-width: 900px) {
    .panes.split {
      grid-template-columns: 1fr;
      grid-template-rows: 1fr 1fr;
      height: auto;
    }
    .split textarea {
      min-height: 24rem;
      border-right: none;
      border-bottom: 1px solid var(--border);
    }
  }
</style>
