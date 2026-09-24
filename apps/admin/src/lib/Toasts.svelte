<script lang="ts">
  import { toasts } from './toast.svelte'
</script>

<div class="toasts" aria-live="polite">
  {#each toasts.items as t (t.id)}
    <button class={['toast', t.kind]} onclick={() => toasts.dismiss(t.id)}>
      <span class="dot"></span>{t.message}
    </button>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    right: 1rem;
    bottom: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    z-index: 100;
  }
  .toast {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font: inherit;
    text-align: left;
    padding: 0.7rem 1rem;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--text);
    box-shadow: 0 8px 30px rgb(0 0 0 / 0.15);
    cursor: pointer;
    max-width: 24rem;
    animation: in 0.18s ease-out;
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex: none;
    background: var(--ok);
  }
  .error .dot {
    background: var(--danger);
  }
  @keyframes in {
    from {
      transform: translateY(8px);
      opacity: 0;
    }
  }
</style>
