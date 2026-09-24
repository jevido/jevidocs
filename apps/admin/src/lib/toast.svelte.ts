export type Toast = { id: number; kind: 'ok' | 'error'; message: string }

let next = 1

class Toasts {
  items = $state<Toast[]>([])

  push(kind: Toast['kind'], message: string) {
    const id = next++
    this.items.push({ id, kind, message })
    setTimeout(() => this.dismiss(id), kind === 'error' ? 6000 : 3000)
  }

  dismiss(id: number) {
    this.items = this.items.filter((t) => t.id !== id)
  }
}

export const toasts = new Toasts()

export const toast = {
  ok: (m: string) => toasts.push('ok', m),
  error: (e: unknown) => toasts.push('error', e instanceof Error ? e.message : String(e)),
}
