<script lang="ts">
  import { api, session } from '../lib/api.svelte'

  let email = $state('')
  let password = $state('')
  let error = $state('')
  let busy = $state(false)

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    error = ''
    busy = true
    try {
      const r = await api.login(email, password)
      session.set(r.token, r.user)
    } catch (err) {
      error = err instanceof Error ? err.message : 'Sign-in failed'
    } finally {
      busy = false
    }
  }
</script>

<div class="wrap">
  <form class="card" onsubmit={submit}>
    <div class="head">
      <img src="/favicon.svg" alt="" width="36" height="36" />
      <h1>Sign in to jevidocs</h1>
      <p class="muted">Manage your documentation projects.</p>
    </div>
    <label class="field">
      Email
      <input type="email" autocomplete="username" required bind:value={email} />
    </label>
    <label class="field">
      Password
      <input type="password" autocomplete="current-password" required bind:value={password} />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn primary" type="submit" disabled={busy}>{busy ? 'Signing in…' : 'Sign in'}</button>
  </form>
</div>

<style>
  .wrap {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 1rem;
    background:
      radial-gradient(60rem 30rem at 50% -10%, color-mix(in srgb, var(--accent) 12%, transparent), transparent),
      var(--bg);
  }
  form {
    width: min(24rem, 100%);
    padding: 2rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .head {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    margin-bottom: 0.5rem;
  }
  .head h1 {
    font-size: 1.25rem;
  }
  .head p {
    margin: 0;
  }
  .error {
    margin: 0;
    color: var(--danger);
    font-size: 0.85rem;
  }
  .btn {
    height: 2.4rem;
  }
</style>
