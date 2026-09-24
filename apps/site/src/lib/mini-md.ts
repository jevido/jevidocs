// Minimal, safe Markdown for AI answers: everything is escaped first, then
// paragraphs, fenced code, inline code, bold and http(s) links are added
// back. Nothing else from the answer reaches the DOM as HTML.
const esc = (s: string) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')

function inline(s: string): string {
  return esc(s)
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noreferrer">$1</a>')
}

export function miniMarkdown(src: string): string {
  const out: string[] = []
  const parts = src.split(/```[^\n]*\n?/)
  parts.forEach((part, i) => {
    if (i % 2 === 1) {
      out.push(`<pre><code>${esc(part.replace(/\n$/, ''))}</code></pre>`)
      return
    }
    for (const para of part.split(/\n{2,}/)) {
      const text = para.trim()
      if (!text) continue
      const lines = text.split('\n')
      if (lines.every((l) => /^\s*[-*] /.test(l))) {
        out.push('<ul>' + lines.map((l) => `<li>${inline(l.replace(/^\s*[-*] /, ''))}</li>`).join('') + '</ul>')
      } else {
        out.push(`<p>${lines.map(inline).join('<br>')}</p>`)
      }
    }
  })
  return out.join('')
}
