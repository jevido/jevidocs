import { session } from './api.svelte'

// Mirrors store.RequiredRole in the API, which is what actually enforces
// it; the UI only hides what the current user cannot do.
export type Role = 'viewer' | 'editor' | 'admin'
const rank: Record<Role, number> = { viewer: 1, editor: 2, admin: 3 }

export const roles: Role[] = ['viewer', 'editor', 'admin']

export function roleOf(): Role {
  const r = session.user?.role as Role | undefined
  return r && rank[r] ? r : 'admin'
}

export function can(min: Role): boolean {
  return rank[roleOf()] >= rank[min]
}
