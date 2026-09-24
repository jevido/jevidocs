<script lang="ts">
  import { session } from '../lib/api.svelte'
  import { assetsApi, type AdminUser } from '../lib/assets-api'
  import { formatDate } from '../lib/format'
  import { toast } from '../lib/toast.svelte'
  import { can, roles } from '../lib/roles'

  let users = $state.raw<AdminUser[]>([])
  let loading = $state(true)
  let draft = $state({ name: '', email: '', password: '', role: 'editor' })
  let adding = $state(false)
  let busy = $state(false)
  let pw = $state({ current: '', next: '', confirm: '' })
  let pwBusy = $state(false)

  function load() {
    assetsApi
      .users()
      .then((u) => (users = u))
      .catch(toast.error)
      .finally(() => (loading = false))
  }
  load()

  async function add(e: SubmitEvent) {
    e.preventDefault()
    busy = true
    try {
      await assetsApi.createUser(draft)
      toast.ok(`Added ${draft.email}`)
      draft = { name: '', email: '', password: '', role: 'editor' }
      adding = false
      load()
    } catch (err) {
      toast.error(err)
    } finally {
      busy = false
    }
  }

  async function setRole(u: AdminUser, role: string) {
    try {
      await assetsApi.setRole(u.id, role)
      toast.ok(`${u.email} is now ${role}`)
    } catch (err) {
      toast.error(err)
    }
    load()
  }

  async function remove(u: AdminUser) {
    if (!confirm(`Delete ${u.email}? Their sessions and API tokens stop working.`)) return
    try {
      await assetsApi.deleteUser(u.id)
      users = users.filter((x) => x.id !== u.id)
      toast.ok('User deleted')
    } catch (err) {
      toast.error(err)
    }
  }

  async function changePassword(e: SubmitEvent) {
    e.preventDefault()
    if (pw.next !== pw.confirm) {
      toast.error('The new passwords do not match')
      return
    }
    pwBusy = true
    try {
      await assetsApi.changePassword(pw.current, pw.next)
      pw = { current: '', next: '', confirm: '' }
      toast.ok('Password changed')
    } catch (err) {
      toast.error(err)
    } finally {
      pwBusy = false
    }
  }
</script>

<div class="page-head">
  <div>
    <h1>Users</h1>
    <p>Everyone listed here can sign in to the admin and edit every project.</p>
  </div>
  {#if can('admin')}
    <button class="btn primary" onclick={() => (adding = !adding)}>{adding ? 'Cancel' : '+ Add user'}</button>
  {/if}
</div>

{#if adding}
  <form class="card card-pad form" onsubmit={add}>
    <label class="field">Name <input type="text" bind:value={draft.name} placeholder="Optional" /></label>
    <label class="field">Email <input type="email" required bind:value={draft.email} /></label>
    <label class="field">
      Password <input type="password" required minlength="8" autocomplete="new-password" bind:value={draft.password} />
    </label>
    <label class="field">
      Role
      <select bind:value={draft.role}>
        {#each roles as r (r)}<option value={r}>{r}</option>{/each}
      </select>
    </label>
    <div><button class="btn primary" type="submit" disabled={busy}>{busy ? 'Adding…' : 'Add user'}</button></div>
  </form>
{/if}

<div class="card">
  {#if users.length === 0}
    {#if !loading}<div class="empty">No users.</div>{/if}
  {:else}
    <table class="list">
      <thead><tr><th>Name</th><th>Email</th><th>Role</th><th>Added</th><th><span class="sr-only">Actions</span></th></tr></thead>
      <tbody>
        {#each users as u (u.id)}
          <tr>
            <td><strong>{u.name}</strong>{#if u.id === session.user?.id} <span class="badge">you</span>{/if}</td>
            <td class="muted">{u.email}</td>
            <td>
              {#if can('admin')}
                <select aria-label="Role of {u.email}" value={u.role} onchange={(e) => setRole(u, e.currentTarget.value)}>
                  {#each roles as r (r)}<option value={r}>{r}</option>{/each}
                </select>
              {:else}
                <span class="badge">{u.role}</span>
              {/if}
            </td>
            <td class="muted">{formatDate(u.created_at)}</td>
            <td class="right">
              {#if u.id !== session.user?.id}
                {#if can('admin')}<button class="btn sm ghost danger" onclick={() => remove(u)}>Delete</button>{/if}
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<h2>Change your password</h2>
<form class="card card-pad form" onsubmit={changePassword}>
  <label class="field">
    Current password <input type="password" required autocomplete="current-password" bind:value={pw.current} />
  </label>
  <label class="field">
    New password <input type="password" required minlength="8" autocomplete="new-password" bind:value={pw.next} />
  </label>
  <label class="field">
    Repeat new password <input type="password" required minlength="8" autocomplete="new-password" bind:value={pw.confirm} />
  </label>
  <div><button class="btn" type="submit" disabled={pwBusy}>{pwBusy ? 'Saving…' : 'Change password'}</button></div>
</form>

<style>
  .form {
    display: grid;
    gap: 0.75rem;
    max-width: 28rem;
    margin-bottom: 1rem;
  }
  h2 { font-size: 1.05rem; margin: 2rem 0 0.75rem; }
</style>
