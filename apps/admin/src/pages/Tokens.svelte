<script lang="ts">
  import { api, API_URL } from '../lib/api.svelte'
  import { formatDate } from '../lib/format'
  import { toast } from '../lib/toast.svelte'
  import type { Token } from '../lib/types'

  let tokens = $state.raw<Token[]>([])
  let loading = $state(true)
  let name = $state('')
  let busy = $state(false)
  let created = $state<{ name: string; token: string } | null>(null)
  let copied = $state(false)

  function load() {
    api
      .tokens()
      .then((t) => (tokens = t))
      .catch(toast.error)
      .finally(() => (loading = false))
  }
  load()

  async function create(e: SubmitEvent) {
    e.preventDefault()
    if (!name.trim()) return
    busy = true
    try {
      const t = await api.createToken(name.trim())
      created = { name: t.name, token: t.token }
      copied = false
      name = ''
      load()
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  async function revoke(t: Token) {
    if (!confirm(`Revoke "${t.name}"? Clients using it stop working immediately.`)) return
    try {
      await api.deleteToken(t.id)
      tokens = tokens.filter((x) => x.id !== t.id)
      toast.ok('Token revoked')
    } catch (err) {
      toast.error(err)
    }
  }

  async function copy() {
    if (!created) return
    try {
      await navigator.clipboard.writeText(created.token)
      copied = true
    } catch {
      toast.error('Copy failed; select the token and copy it by hand')
    }
  }

  const mcpConfig = $derived(
    JSON.stringify(
      {
        mcpServers: {
          jevidocs: {
            type: 'http',
            url: `${API_URL}/mcp`,
            headers: { Authorization: `Bearer ${created?.token ?? '<token>'}` },
          },
        },
      },
      null,
      2,
    ),
  )
</script>

<div class="page-head">
  <div>
    <h1>API tokens</h1>
    <p>
      Tokens let MCP clients and scripts write docs. Send them as
      <code>Authorization: Bearer &lt;token&gt;</code> to <code>{API_URL}/mcp</code> or the admin API.
    </p>
  </div>
</div>

<form class="card card-pad create" onsubmit={create}>
  <label class="field">
    Token name
    <input type="text" placeholder="Claude Code on my laptop" bind:value={name} />
  </label>
  <button class="btn primary" type="submit" disabled={busy || !name.trim()}>Create token</button>
</form>

{#if created}
  <div class="card card-pad fresh">
    <div class="row">
      <strong>{created.name}</strong>
      <span class="badge warn">Shown once</span>
    </div>
    <p class="muted">Copy it now. It cannot be shown again.</p>
    <div class="row">
      <code class="token">{created.token}</code>
      <button class="btn" onclick={copy}>{copied ? 'Copied ✓' : 'Copy'}</button>
    </div>
    <p class="muted">MCP client config, for example <code>.mcp.json</code>:</p>
    <pre>{mcpConfig}</pre>
    <div><button class="btn sm ghost" onclick={() => (created = null)}>Done</button></div>
  </div>
{/if}

<div class="card">
  {#if tokens.length === 0}
    {#if !loading}<div class="empty">No tokens yet.</div>{/if}
  {:else}
    <table class="list">
      <thead><tr><th>Name</th><th>Created</th><th>Last used</th><th><span class="sr-only">Actions</span></th></tr></thead>
      <tbody>
        {#each tokens as t (t.id)}
          <tr>
            <td><strong>{t.name}</strong></td>
            <td class="muted">{formatDate(t.created_at)}</td>
            <td class="muted">{t.last_used_at ? formatDate(t.last_used_at) : 'Never'}</td>
            <td class="right"><button class="btn sm ghost danger" onclick={() => revoke(t)}>Revoke</button></td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .page-head p {
    max-width: 44rem;
  }
  .create {
    display: flex;
    align-items: flex-end;
    gap: 0.75rem;
    margin-bottom: 1rem;
    max-width: 40rem;
  }
  .create .field {
    flex: 1;
  }
  .fresh {
    margin-bottom: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    border-color: color-mix(in srgb, var(--warn) 45%, var(--border));
  }
  .fresh p {
    margin: 0;
  }
  .token {
    flex: 1;
    padding: 0.5rem 0.7rem;
    border-radius: 8px;
    background: var(--surface-2);
    border: 1px solid var(--border);
    overflow-x: auto;
    white-space: nowrap;
    user-select: all;
  }
  pre {
    margin: 0;
    padding: 0.8rem;
    border-radius: 8px;
    background: var(--surface-2);
    border: 1px solid var(--border);
    font-family: var(--mono);
    font-size: 0.8rem;
    overflow-x: auto;
  }
  .right {
    text-align: right;
  }
</style>
