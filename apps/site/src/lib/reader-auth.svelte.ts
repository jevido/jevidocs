// Who is reading: an optional signed-in user (for private projects they may
// read) and per-project share tokens from ?share= links. Both live in
// localStorage, wrapped in try/catch because storage can be blocked.
import { API_URL } from './api-base'

const TOKEN_KEY = 'jevidocs-reader-token'
const USER_KEY = 'jevidocs-reader-user'
const shareKey = (project: string) => `jevidocs-share:${project}`

export type ReaderUser = { id: number; name: string; email: string; role?: string }

function read(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

function write(key: string, value: string | null) {
  try {
    if (value === null) localStorage.removeItem(key)
    else localStorage.setItem(key, value)
  } catch {
    // storage blocked: the session lasts until reload
  }
}

function readUser(): ReaderUser | null {
  const raw = read(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as ReaderUser
  } catch {
    return null
  }
}

class ReaderAuth {
  token = $state<string | null>(read(TOKEN_KEY))
  user = $state<ReaderUser | null>(readUser())
  // Bumped on sign in/out so views depending on access reload.
  version = $state(0)
  // Share tokens kept in memory too, in case storage is blocked.
  #shares = new Map<string, string>()

  async signIn(email: string, password: string): Promise<void> {
    const res = await fetch(`${API_URL}/api/auth/login`, {
      method: 'POST',
      headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok || !data?.token) throw new Error(data?.error ?? 'Could not sign in')
    this.token = data.token
    this.user = data.user ?? null
    write(TOKEN_KEY, data.token)
    write(USER_KEY, data.user ? JSON.stringify(data.user) : null)
    this.version++
  }

  signOut() {
    const token = this.token
    this.token = null
    this.user = null
    write(TOKEN_KEY, null)
    write(USER_KEY, null)
    this.version++
    if (token) {
      void fetch(`${API_URL}/api/auth/logout`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {})
    }
  }

  // A stale or revoked token: forget it, but keep reading as anonymous.
  expire() {
    if (!this.token) return
    this.token = null
    this.user = null
    write(TOKEN_KEY, null)
    write(USER_KEY, null)
    this.version++
  }

  share(project: string): string | null {
    return this.#shares.get(project) ?? read(shareKey(project))
  }

  // Takes ?share=<token> off the current URL and remembers it for project.
  captureShare(project: string) {
    if (!project) return
    try {
      const url = new URL(location.href)
      const token = url.searchParams.get('share')
      if (!token) return
      this.#shares.set(project, token)
      write(shareKey(project), token)
      url.searchParams.delete('share')
      history.replaceState(history.state, '', url.pathname + url.search + url.hash)
    } catch {
      // malformed URL: ignore
    }
  }

  forgetShare(project: string) {
    this.#shares.delete(project)
    write(shareKey(project), null)
  }

  headers(project?: string): Record<string, string> {
    const h: Record<string, string> = {}
    if (this.token) h.Authorization = `Bearer ${this.token}`
    const s = project ? this.share(project) : null
    if (s) h['X-Jevidocs-Share'] = s
    return h
  }
}

export const readerAuth = new ReaderAuth()
