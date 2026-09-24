<script lang="ts">
  import { accessApi, type Member, type NewShareLink, type ShareLink } from './access-api'
  import { assetsApi, type AdminUser } from './assets-api'
  import { formatDate } from './format'
  import { can } from './roles'
  import { toast } from './toast.svelte'

  let { project, isPublic }: { project: string; isPublic: boolean } = $props()

  let members = $state.raw<Member[]>([])
  let users = $state.raw<AdminUser[]>([])
  let shares = $state.raw<ShareLink[]>([])
  let loading = $state(true)
  let pick = $state('')
  let shareName = $state('')
  let shareDays = $state(30)
  let created = $state<NewShareLink | null>(null)
  let busy = $state(false)

  const admin = can('admin')
  // Admins read every private project already; members matter for the rest.
  const candidates = $derived(
    users.filter((u) => u.role !== 'admin' && !members.some((m) => m.user_id === u.id)),
  )

  function load() {
    loading = true
    Promise.all([
      accessApi.members(project).then((m) => (members = m ?? [])),
      accessApi.shares(project).then((s) => (shares = s ?? [])),
      admin ? assetsApi.users().then((u) => (users = u ?? [])) : Promise.resolve(),
    ])
      .catch(toast.error)
      .finally(() => (loading = false))
  }
  load()

  async function addMember(e: SubmitEvent) {
    e.preventDefault()
    const id = Number(pick)
    if (!id) return
    busy = true
    try {
      await accessApi.addMember(project, id)
      pick = ''
      members = await accessApi.members(project)
      toast.ok('Member added')
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  async function removeMember(m: Member) {
    if (!confirm(`Remove ${m.name} from this project?`)) return
    try {
      await accessApi.removeMember(project, m.user_id)
      members = members.filter((x) => x.user_id !== m.user_id)
      toast.ok('Member removed')
    } catch (err) {
      toast.error(err)
    }
  }

  async function createShare(e: SubmitEvent) {
    e.preventDefault()
    if (!shareName.trim()) return
    busy = true
    try {
      created = await accessApi.createShare(project, shareName.trim(), shareDays)
      shareName = ''
      shares = await accessApi.shares(project)
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  async function revoke(s: ShareLink) {
    if (!confirm(`Revoke "${s.name}"? Anyone using this link loses access.`)) return
    try {
      await accessApi.revokeShare(project, s.id)
      shares = shares.filter((x) => x.id !== s.id)
      if (created?.id === s.id) created = null
      toast.ok('Link revoked')
    } catch (err) {
      toast.error(err)
    }
  }

  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text)
      toast.ok('Copied')
    } catch {
      toast.error('Copy failed; select the link and copy it by hand')
    }
  }

  const expired = (s: ShareLink) => !!s.expires_at && new Date(s.expires_at).getTime() < Date.now()
</script>

<section class="card card-pad access">
  <h2>Who can read this project</h2>
  {#if isPublic}
    <p class="note">
      This project is <strong>public</strong>: anyone can read it on the docs site, in search and over MCP. Untick
      <em>Public</em> in the settings above to make it private; the members and share links below apply from then on.
    </p>
  {:else}
    <p class="note">
      This project is <strong>private</strong>. It reads as “not found” for everyone except: admins, the members below
      (after signing in on the docs site with their jevidocs account), and anyone with a share link.
    </p>
  {/if}

  <h3>Members</h3>
  {#if loading}
    <p class="muted">Loading…</p>
  {:else}
    {#if members.length === 0}
      <p class="muted">No members yet. Admins can always read private projects.</p>
    {:else}
      <table>
        <thead><tr><th>Name</th><th>Email</th><th>Role</th><th><span class="sr-only">Actions</span></th></tr></thead>
        <tbody>
          {#each members as m (m.user_id)}
            <tr>
              <td>{m.name}</td>
              <td>{m.email}</td>
              <td><span class="badge">{m.role}</span></td>
              <td class="end">
                {#if admin}<button class="btn sm danger" type="button" onclick={() => removeMember(m)}>Remove</button>{/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
    {#if admin}
      <form class="row" onsubmit={addMember}>
        <label class="sr-only" for="member-pick">User to add</label>
        <select id="member-pick" bind:value={pick} disabled={candidates.length === 0}>
          <option value="">{candidates.length ? 'Choose a user…' : 'No other editors or viewers'}</option>
          {#each candidates as u (u.id)}
            <option value={String(u.id)}>{u.name} ({u.email}, {u.role})</option>
          {/each}
        </select>
        <button class="btn" type="submit" disabled={!pick || busy}>Add member</button>
        <a class="hint" href="#/users">Create users on the Users page</a>
      </form>
    {/if}
  {/if}

  <h3>Share links</h3>
  <p class="muted">Anyone with a link can read the whole project without an account, until it expires or is revoked.</p>
  {#if created}
    <div class="created" role="status">
      <p><strong>{created.name}</strong>: copy this link now, it is shown only once.</p>
      <div class="row">
        <input class="mono" readonly value={created.url} onfocus={(e) => (e.currentTarget as HTMLInputElement).select()} />
        <button class="btn primary" type="button" onclick={() => copy(created!.url)}>Copy</button>
        <button class="btn" type="button" onclick={() => (created = null)}>Done</button>
      </div>
    </div>
  {/if}
  {#if shares.length > 0}
    <table>
      <thead><tr><th>Name</th><th>Expires</th><th>Last used</th><th>Created</th><th><span class="sr-only">Actions</span></th></tr></thead>
      <tbody>
        {#each shares as s (s.id)}
          <tr class:expired={expired(s)}>
            <td>{s.name}</td>
            <td>{s.expires_at ? (expired(s) ? 'Expired ' : '') + formatDate(s.expires_at) : 'Never'}</td>
            <td>{s.last_used_at ? formatDate(s.last_used_at) : '—'}</td>
            <td>{formatDate(s.created_at)}</td>
            <td class="end">
              {#if admin}<button class="btn sm danger" type="button" onclick={() => revoke(s)}>Revoke</button>{/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if !loading}
    <p class="muted">No share links.</p>
  {/if}
  {#if admin}
    <form class="row" onsubmit={createShare}>
      <label class="sr-only" for="share-name">Link name</label>
      <input id="share-name" type="text" placeholder="Name, e.g. Acme team" bind:value={shareName} maxlength="100" />
      <label class="sr-only" for="share-days">Expires</label>
      <select id="share-days" bind:value={shareDays}>
        <option value={7}>Expires in 7 days</option>
        <option value={30}>Expires in 30 days</option>
        <option value={90}>Expires in 90 days</option>
        <option value={0}>Never expires</option>
      </select>
      <button class="btn" type="submit" disabled={!shareName.trim() || busy}>Create link</button>
    </form>
  {:else}
    <p class="muted">Only admins manage members and share links.</p>
  {/if}
</section>

<style>
  .access { display: grid; gap: 0.6rem; margin-top: 1rem; }
  h2 { margin: 0; font-size: 1.05rem; }
  h3 { margin: 0.75rem 0 0; font-size: 0.95rem; }
  .note { margin: 0; font-size: 0.9rem; }
  .muted { margin: 0; color: var(--muted); font-size: 0.85rem; }
  .row { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
  .row input, .row select { flex: 1 1 14rem; min-width: 0; }
  .hint { font-size: 0.85rem; color: var(--muted); }
  table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
  th, td { text-align: left; padding: 0.45rem 0.5rem; border-bottom: 1px solid var(--border); }
  th { font-weight: 500; color: var(--muted); }
  td.end { text-align: right; }
  tr.expired td { color: var(--muted); }
  .created {
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    background: var(--surface-2);
    display: grid;
    gap: 0.5rem;
  }
  .created p { margin: 0; font-size: 0.875rem; }
  .mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.8rem; }
  .sr-only {
    position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px;
    overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0;
  }
</style>
