<script lang="ts">
  let { method, small = false }: { method: string; small?: boolean } = $props()
</script>

<span class="method" class:small style:--c="var(--method-{method.toLowerCase()}, var(--muted-fg))">{method.toUpperCase()}</span>

<style>
  .method {
    flex: none;
    display: inline-block;
    padding: 0.05rem 0.4rem;
    border-radius: 5px;
    font: 600 0.7rem/1.5 var(--font-mono);
    letter-spacing: 0.02em;
    color: var(--c);
    background: color-mix(in oklab, var(--c) 14%, transparent);
  }
  .small {
    padding: 0;
    background: none;
    font-size: 0.62rem;
  }
</style>
