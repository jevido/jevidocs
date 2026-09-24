<script lang="ts">
  import './preview.css'

  let { html }: { html: string } = $props()

  // Tabs and code-copy behaviour for the server-rendered HTML. Delegated on
  // the container, so it survives every {@html} update.
  function behaviour(node: HTMLElement) {
    function onclick(e: MouseEvent) {
      const target = e.target as HTMLElement
      const trigger = target.closest<HTMLElement>('.fd-tab-trigger')
      if (trigger) {
        const tabs = trigger.closest('.fd-tabs')
        const value = trigger.dataset.tab
        if (!tabs || value === undefined) return
        tabs.querySelectorAll<HTMLElement>(':scope > .fd-tabs-list > .fd-tab-trigger').forEach((b) => {
          b.toggleAttribute('data-active', b.dataset.tab === value)
        })
        tabs.querySelectorAll<HTMLElement>(':scope > .fd-tab').forEach((t) => {
          t.toggleAttribute('data-active', t.dataset.value === value)
        })
        return
      }
      const copy = target.closest<HTMLElement>('.fd-copy')
      if (copy) {
        const code = copy.closest('.fd-codeblock')?.querySelector('code')?.innerText ?? ''
        navigator.clipboard?.writeText(code).then(() => {
          copy.textContent = 'Copied'
          setTimeout(() => (copy.textContent = 'Copy'), 1200)
        })
      }
    }
    node.addEventListener('click', onclick)
    return () => node.removeEventListener('click', onclick)
  }

  function withCopyButtons(src: string): string {
    return src.replace(/<figure class="fd-codeblock"([^>]*)>/g, '<figure class="fd-codeblock"$1><button type="button" class="fd-copy">Copy</button>')
  }
</script>

<div class="fd-prose" {@attach behaviour}>
  {@html withCopyButtons(html)}
</div>
