import type { AdminPageInput } from './types'

// A page handed from "Translate" to the new-page editor, which starts from
// it instead of the empty template. Read once.
let pending: AdminPageInput | null = null

export function startTranslation(page: AdminPageInput) {
  pending = { ...page }
}

export function takeTranslation(): AdminPageInput | null {
  const p = pending
  pending = null
  return p
}
