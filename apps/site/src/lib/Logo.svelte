<script lang="ts">
  let { size = 22 }: { size?: number } = $props()
</script>

<svg width={size} height={size} viewBox="0 0 32 32" aria-hidden="true">
  <defs>
    <linearGradient id="jd-logo" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0" stop-color="#a78bfa" />
      <stop offset="1" stop-color="#38bdf8" />
    </linearGradient>
  </defs>
  <path d="M6 5h13a7 7 0 0 1 7 7v15H13a7 7 0 0 1-7-7z" fill="url(#jd-logo)" />
  <path d="M12 12h8M12 17h8M12 22h4" stroke="var(--bg)" stroke-width="2.4" stroke-linecap="round" />
</svg>
