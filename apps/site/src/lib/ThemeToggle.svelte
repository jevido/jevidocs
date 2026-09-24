<script lang="ts">
  import Icon from './Icon.svelte'
  import { theme, type ThemeMode } from './theme.svelte'

  const modes: { mode: ThemeMode; icon: string; label: string }[] = [
    { mode: 'light', icon: 'sun', label: 'Light' },
    { mode: 'dark', icon: 'moon', label: 'Dark' },
    { mode: 'system', icon: 'monitor', label: 'System' },
  ]
</script>

<div class="toggle" role="radiogroup" aria-label="Theme">
  {#each modes as m (m.mode)}
    <button
      type="button"
      role="radio"
      aria-checked={theme.mode === m.mode}
      aria-label={m.label}
      title={m.label}
      class:active={theme.mode === m.mode}
      onclick={() => theme.set(m.mode)}>
      <Icon name={m.icon} size={14} />
    </button>
  {/each}
</div>

<style>
  .toggle {
    display: inline-flex;
    gap: 2px;
    padding: 2px;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--card);
  }
  button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.6rem;
    height: 1.6rem;
    border: 0;
    border-radius: 999px;
    background: transparent;
    color: var(--muted-fg);
    cursor: pointer;
  }
  button:hover { color: var(--fg); }
  button.active {
    background: var(--accent);
    color: var(--fg);
  }
</style>
