// Feedback and page history endpoints (see services/api/README.md).
import { request } from './api.svelte'
import type { AdminPage } from './types'

const enc = encodeURIComponent

export type FeedbackEntry = { id: number; slug: string; helpful: boolean; message: string; created_at: string }
export type FeedbackTotal = { slug: string; helpful: number; not_helpful: number }
export type RevisionSummary = { id: number; title: string; created_at: string; size: number }
export type Revision = RevisionSummary & { description: string; body: string }

export const extraApi = {
  feedback: (project: string) =>
    request<{ entries: FeedbackEntry[]; totals: FeedbackTotal[] }>('GET', `/api/admin/projects/${enc(project)}/feedback`),
  revisions: (project: string, page: number) =>
    request<RevisionSummary[]>('GET', `/api/admin/projects/${enc(project)}/pages/${page}/revisions`),
  revision: (project: string, page: number, id: number) =>
    request<Revision>('GET', `/api/admin/projects/${enc(project)}/pages/${page}/revisions/${id}`),
  restore: (project: string, page: number, id: number) =>
    request<AdminPage>('POST', `/api/admin/projects/${enc(project)}/pages/${page}/revisions/${id}/restore`),
}

export type DiffLine = { kind: 'same' | 'add' | 'del'; text: string }

// lineDiff is a longest-common-subsequence diff of two texts by line, from
// `from` (the revision) to `to` (the current page). Big inputs fall back to
// "all removed, all added" rather than an O(n·m) table.
export function lineDiff(from: string, to: string): DiffLine[] {
  const a = from.split('\n')
  const b = to.split('\n')
  if (a.length * b.length > 4_000_000) {
    return [...a.map((text) => ({ kind: 'del' as const, text })), ...b.map((text) => ({ kind: 'add' as const, text }))]
  }
  const n = a.length
  const m = b.length
  const lcs: Uint32Array[] = Array.from({ length: n + 1 }, () => new Uint32Array(m + 1))
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      lcs[i][j] = a[i] === b[j] ? lcs[i + 1][j + 1] + 1 : Math.max(lcs[i + 1][j], lcs[i][j + 1])
    }
  }
  const out: DiffLine[] = []
  let i = 0
  let j = 0
  while (i < n && j < m) {
    if (a[i] === b[j]) {
      out.push({ kind: 'same', text: a[i] })
      i++
      j++
    } else if (lcs[i + 1][j] >= lcs[i][j + 1]) {
      out.push({ kind: 'del', text: a[i++] })
    } else {
      out.push({ kind: 'add', text: b[j++] })
    }
  }
  while (i < n) out.push({ kind: 'del', text: a[i++] })
  while (j < m) out.push({ kind: 'add', text: b[j++] })
  return out
}
