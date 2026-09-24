// Paste or drop files into the Markdown textarea: each file gets an
// "Uploading…" placeholder at the cursor, replaced by its Markdown once the
// upload finishes (or removed if it fails).
import { assetsApi, assetMarkdown } from './assets-api'
import { toast } from './toast.svelte'

export function filesFrom(e: ClipboardEvent | DragEvent): File[] {
  const list = 'clipboardData' in e ? e.clipboardData?.files : (e as DragEvent).dataTransfer?.files
  return list ? Array.from(list) : []
}

export async function uploadInto(
  el: HTMLTextAreaElement,
  project: string,
  files: File[],
  setBody: (v: string) => void,
): Promise<void> {
  for (const file of files) {
    const placeholder = `![Uploading ${file.name}…]()`
    el.setRangeText(placeholder, el.selectionStart, el.selectionEnd, 'end')
    setBody(el.value)
    let replacement = ''
    try {
      const a = await assetsApi.upload(project, file)
      replacement = assetMarkdown(a)
    } catch (err) {
      toast.error(err)
    }
    const at = el.value.indexOf(placeholder)
    if (at >= 0) {
      el.setRangeText(replacement, at, at + placeholder.length, 'preserve')
      setBody(el.value)
    }
  }
}
