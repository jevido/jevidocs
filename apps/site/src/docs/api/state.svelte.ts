import { createContext } from 'svelte'
import { serverURL, type ApiOperation, type ApiReference, type AuthValues } from '../../lib/openapi'

const CLIENT_KEY = 'jevidocs-api-client'

function read<T>(store: () => Storage, key: string, fallback: T): T {
  try {
    const v = store().getItem(key)
    return v === null ? fallback : (JSON.parse(v) as T)
  } catch {
    return fallback
  }
}

function write(store: () => Storage, key: string, v: unknown) {
  try {
    store().setItem(key, JSON.stringify(v))
  } catch {
    // storage unavailable: keep it for this page view only
  }
}

// ApiState is what the reader chose on an OpenAPI page: server, credentials
// and code sample language. Credentials stay in sessionStorage (this tab
// only); the language is remembered across pages.
export class ApiState {
  readonly ref: ApiReference
  readonly #authKey: string
  serverIndex = $state(0)
  serverVars = $state<Record<string, string>>({})
  customServer = $state('')
  auth = $state<AuthValues>({})
  scheme = $state('')
  client = $state(read(() => localStorage, CLIENT_KEY, 'curl'))
  // The operation open in the request client, if any.
  testing = $state.raw<ApiOperation | null>(null)
  server = $derived.by(() =>
    this.ref.servers.length ? serverURL(this.ref.servers[this.serverIndex], this.serverVars) : this.customServer.replace(/\/+$/, ''),
  )

  constructor(ref: ApiReference, project: string) {
    this.ref = ref
    this.#authKey = `jevidocs-api-auth:${project}`
    this.auth = read(() => sessionStorage, this.#authKey, {})
    this.scheme = ref.security_schemes[0]?.key ?? ''
    for (const v of ref.servers[0]?.variables ?? []) this.serverVars[v.name] = v.default
  }

  setClient(id: string) {
    this.client = id
    write(() => localStorage, CLIENT_KEY, id)
  }

  setAuth(key: string, field: 'token' | 'username' | 'password' | 'value', value: string) {
    this.auth[key] = { ...this.auth[key], [field]: value }
    write(() => sessionStorage, this.#authKey, $state.snapshot(this.auth))
  }
}

export const [getApi, setApi] = createContext<ApiState>()

// apiNav hands the open OpenAPI page's outline to the sidebar, which lives
// outside the page, and takes back the operation in view.
class ApiNav {
  slug = $state('')
  ref = $state.raw<ApiReference | null>(null)
  active = $state('')
}

export const apiNav = new ApiNav()
