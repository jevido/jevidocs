<script lang="ts" module>
  // One dialog for the whole page, opened from the navbar, the project
  // directory or a not-found state.
  let dialogOpen = $state(false)
  export function openSignIn() {
    dialogOpen = true
  }
</script>

<script lang="ts">
  import { tick } from 'svelte'
  import { readerAuth } from '../lib/reader-auth.svelte'

  // `compact` renders only the dialog (no navbar button), for pages that
  // just need openSignIn().
  let { compact = false }: { compact?: boolean } = $props()

  let email = $state('')
  let password = $state('')
  let busy = $state(false)
  let error = $state('')
  let menuOpen = $state(false)
  let dialog = $state<HTMLDialogElement | null>(null)
  let emailInput = $state<HTMLInputElement | null>(null)

  $effect(() => {
    if (!dialog) return
    if (dialogOpen && !dialog.open) {
      error = ''
      dialog.showModal()
      void tick().then(() => emailInput?.focus())
    } else if (!dialogOpen && dialog.open) {
      dialog.close()
    }
  })

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    busy = true
    error = ''
    try {
      await readerAuth.signIn(email.trim(), password)
      password = ''
      dialogOpen = false
    } catch (err) {
      error = err instanceof Error ? err.message : 'Could not sign in'
    } finally {
      busy = false
    }
  }

  function closeMenu(e: MouseEvent) {
    if (!(e.target as HTMLElement).closest('.account')) menuOpen = false
  }
</script>

<svelte:window onclick={closeMenu} />

{#if !compact}
  {#if readerAuth.user}
    <div class="account">
      <button
        type="button"
        class="account-btn"
        aria-haspopup="menu"
        aria-expanded={menuOpen}
        onclick={() => (menuOpen = !menuOpen)}>
        <span class="avatar" aria-hidden="true">{readerAuth.user.name.slice(0, 1).toUpperCase()}</span>
        <span class="who">{readerAuth.user.name}</span>
      </button>
      {#if menuOpen}
        <div class="menu" role="menu">
          <p class="email">{readerAuth.user.email}</p>
          <button
            type="button"
            role="menuitem"
            onclick={() => {
              menuOpen = false
              readerAuth.signOut()
            }}>Sign out</button>
        </div>
      {/if}
    </div>
  {:else}
    <button type="button" class="signin" onclick={openSignIn}>Sign in</button>
  {/if}
{/if}

<dialog
  bind:this={dialog}
  class="signin-dialog"
  aria-labelledby="signin-title"
  onclose={() => (dialogOpen = false)}
  onclick={(e) => e.target === dialog && (dialogOpen = false)}>
  <form onsubmit={submit}>
    <h2 id="signin-title">Sign in to read private docs</h2>
    <p class="hint">Use the account an admin created for you in jevidocs.</p>
    <label>
      Email
      <input bind:this={emailInput} type="email" autocomplete="username" required bind:value={email} />
    </label>
    <label>
      Password
      <input type="password" autocomplete="current-password" required bind:value={password} />
    </label>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="actions">
      <button type="button" class="ghost" onclick={() => (dialogOpen = false)}>Cancel</button>
      <button type="submit" class="primary" disabled={busy}>{busy ? 'Signing in…' : 'Sign in'}</button>
    </div>
  </form>
</dialog>

<style>
  .signin,
  .account-btn {
    font: inherit;
    font-size: 0.85rem;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg);
    border-radius: 999px;
    padding: 0.3rem 0.8rem;
    cursor: pointer;
    white-space: nowrap;
  }
  .signin:hover,
  .account-btn:hover { background: var(--accent); }
  .account { position: relative; }
  .account-btn { display: inline-flex; align-items: center; gap: 0.45rem; padding: 0.2rem 0.7rem 0.2rem 0.25rem; }
  .avatar {
    display: grid;
    place-items: center;
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 999px;
    background: var(--brand);
    color: white;
    font-size: 0.75rem;
    font-weight: 600;
  }
  .who { max-width: 8rem; overflow: hidden; text-overflow: ellipsis; }
  .menu {
    position: absolute;
    right: 0;
    top: calc(100% + 0.35rem);
    z-index: 40;
    min-width: 12rem;
    padding: 0.35rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--popover);
    box-shadow: var(--shadow);
  }
  .menu .email { margin: 0.25rem 0.5rem 0.4rem; font-size: 0.8rem; color: var(--muted-fg); }
  .menu button {
    width: 100%;
    text-align: left;
    font: inherit;
    font-size: 0.875rem;
    padding: 0.4rem 0.5rem;
    border: 0;
    border-radius: 0.35rem;
    background: transparent;
    color: var(--fg);
    cursor: pointer;
  }
  .menu button:hover { background: var(--accent); }
  .signin-dialog {
    width: min(24rem, calc(100vw - 2rem));
    padding: 0;
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    background: var(--popover);
    color: var(--fg);
    box-shadow: var(--shadow);
  }
  .signin-dialog::backdrop { background: var(--overlay); }
  form { display: grid; gap: 0.75rem; padding: 1.25rem; }
  h2 { margin: 0; font-size: 1.05rem; }
  .hint { margin: 0; font-size: 0.85rem; color: var(--muted-fg); }
  label { display: grid; gap: 0.3rem; font-size: 0.85rem; }
  input {
    font: inherit;
    padding: 0.5rem 0.6rem;
    border: 1px solid var(--border);
    border-radius: 0.4rem;
    background: var(--bg);
    color: var(--fg);
  }
  .error { margin: 0; color: var(--error); font-size: 0.85rem; }
  .actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 0.25rem; }
  .actions button {
    font: inherit;
    font-size: 0.875rem;
    padding: 0.45rem 0.9rem;
    border-radius: 0.4rem;
    cursor: pointer;
    border: 1px solid var(--border);
  }
  .ghost { background: transparent; color: var(--fg); }
  .primary { background: var(--primary); color: var(--primary-fg); border-color: var(--primary); }
  .primary:disabled { opacity: 0.6; cursor: progress; }
</style>
