// Types and helpers for OpenAPI pages: the reference the API derives from a
// spec (services/api/app/openapi, Reference), request building for code
// samples and the request client, and a small highlighter for the code the
// browser generates.

export type ApiSchema = {
  ref?: string
  type?: string
  format?: string
  title?: string
  description?: string
  enum?: unknown[]
  default?: unknown
  example?: unknown
  nullable?: boolean
  read_only?: boolean
  write_only?: boolean
  deprecated?: boolean
  constraints?: string[]
  properties?: ApiProperty[]
  additional_properties?: ApiSchema
  items?: ApiSchema
  composition?: 'oneOf' | 'anyOf' | 'allOf'
  variants?: ApiSchema[]
  truncated?: boolean
}
export type ApiProperty = { name: string; required?: boolean; schema: ApiSchema }
export type ApiExample = { name: string; summary?: string; value: unknown }
export type ApiMedia = { type: string; schema?: ApiSchema; examples?: ApiExample[] }
export type ApiParameter = {
  name: string
  in: 'path' | 'query' | 'header' | 'cookie'
  description?: string
  required?: boolean
  deprecated?: boolean
  schema?: ApiSchema
  example?: unknown
}
export type ApiResponse = { status: string; description?: string; headers?: ApiParameter[]; content?: ApiMedia[] }
export type ApiOperation = {
  id: string
  method: string
  path: string
  operation_id?: string
  summary: string
  description?: string
  deprecated?: boolean
  security: string[]
  parameters: ApiParameter[]
  request_body?: { description?: string; required?: boolean; content: ApiMedia[] }
  responses: ApiResponse[]
  code_samples?: { lang: string; label?: string; source: string }[]
}
export type ApiTag = { id: string; name: string; description?: string; operations: ApiOperation[] }
export type ApiServer = {
  url: string
  description?: string
  variables?: { name: string; default: string; enum?: string[]; description?: string }[]
}
export type ApiSecurityScheme = {
  key: string
  type: string
  scheme?: string
  bearer_format?: string
  in?: string
  name?: string
  description?: string
  openid_url?: string
  flows?: { type: string; authorization_url?: string; token_url?: string; scopes?: { name: string; description?: string }[] }[]
}
export type ApiReference = {
  openapi: string
  title: string
  version: string
  description: string
  servers: ApiServer[]
  security_schemes: ApiSecurityScheme[]
  tags: ApiTag[]
  models: { id: string; name: string }[]
  schemas: Record<string, ApiSchema>
}

// resolve follows named-schema references (guarding against loops).
export function resolve(s: ApiSchema | undefined, schemas: Record<string, ApiSchema>): ApiSchema | undefined {
  for (let i = 0; s?.ref && i < 16; i++) s = schemas[s.ref]
  return s
}

// typeLabel is the short type shown next to a name: "string · email",
// "Pet[]", "object", "integer | null".
export function typeLabel(s: ApiSchema | undefined, schemas: Record<string, ApiSchema>, depth = 0): string {
  if (!s) return 'any'
  if (s.ref) {
    const r = resolve(s, schemas)
    const base = r && r.type && r.type !== 'object' && !r.properties ? typeLabel(r, schemas, depth + 1) : s.ref
    return base
  }
  if (s.variants?.length && depth < 4) {
    const sep = s.composition === 'allOf' ? ' & ' : ' | '
    return s.variants.map((v) => resolve(v, schemas)?.title || typeLabel(v, schemas, depth + 1)).join(sep)
  }
  let t = s.type || (s.properties ? 'object' : 'any')
  if (t === 'array') t = `${s.items ? typeLabel(s.items, schemas, depth + 1) : 'any'}[]`
  if (s.format) t += ` · ${s.format}`
  return t
}

// hasChildren tells whether a schema has attributes worth expanding.
export function hasChildren(s: ApiSchema | undefined, schemas: Record<string, ApiSchema>): boolean {
  const r = resolve(s, schemas)
  if (!r) return false
  if (r.properties?.length || r.variants?.length || r.additional_properties?.properties) return true
  if (r.items) return hasChildren(r.items, schemas)
  return false
}

// ---- requests -----------------------------------------------------------

export type AuthValues = Record<string, { token?: string; username?: string; password?: string; value?: string }>

export type RequestInput = {
  server: string
  path: Record<string, string>
  query: [string, string][]
  headers: [string, string][]
  cookies: [string, string][]
  body?: string
  contentType?: string
}

export type BuiltRequest = { method: string; url: string; headers: [string, string][]; body?: string }

export function serverURL(s: ApiServer | undefined, vars: Record<string, string>): string {
  if (!s) return ''
  return s.url.replace(/\{([^}]+)\}/g, (_, name: string) => vars[name] ?? s.variables?.find((v) => v.name === name)?.default ?? '').replace(/\/+$/, '')
}

function exampleString(v: unknown): string {
  if (v === undefined || v === null) return ''
  return typeof v === 'string' ? v : JSON.stringify(v)
}

// defaultInput fills a request from the spec: examples for parameters, the
// first body example, and placeholders for required values without one.
export function defaultInput(op: ApiOperation, server: string, placeholders = true): RequestInput {
  const path: Record<string, string> = {}
  const query: [string, string][] = []
  const headers: [string, string][] = []
  const cookies: [string, string][] = []
  for (const p of op.parameters) {
    const v = exampleString(p.example ?? p.schema?.example ?? p.schema?.default)
    if (p.in === 'path') path[p.name] = v || (placeholders ? `{${p.name}}` : '')
    else if (p.required) {
      // Optional parameters stay out of samples; the client lists them.
      const pair: [string, string] = [p.name, v || (placeholders ? `{${p.name}}` : '')]
      if (p.in === 'query') query.push(pair)
      else if (p.in === 'header') headers.push(pair)
      else cookies.push(pair)
    }
  }
  const media = op.request_body?.content?.[0]
  let body: string | undefined
  if (media) {
    const ex = media.examples?.[0]?.value
    body = ex === undefined ? '' : typeof ex === 'string' && !media.type.includes('json') ? ex : JSON.stringify(ex, null, 2)
  }
  return { server, path, query, headers, cookies, body, contentType: media?.type }
}

// authHeaders turns the reader's credentials for the schemes an operation
// accepts into headers or query parameters. The first scheme with a value
// wins; without any, the first scheme gets a placeholder.
export function applyAuth(
  op: ApiOperation,
  schemes: ApiSecurityScheme[],
  auth: AuthValues,
  placeholders: boolean,
): { headers: [string, string][]; query: [string, string][]; cookies: [string, string][] } {
  const out = { headers: [] as [string, string][], query: [] as [string, string][], cookies: [] as [string, string][] }
  const usable = op.security.map((k) => schemes.find((s) => s.key === k)).filter((s): s is ApiSecurityScheme => !!s)
  if (!usable.length) return out
  const filled = usable.find((s) => {
    const a = auth[s.key]
    return a && (a.token || a.value || a.username)
  })
  const s = filled ?? (placeholders ? usable[0] : undefined)
  if (!s) return out
  const a = auth[s.key] ?? {}
  if (s.type === 'http' && s.scheme === 'basic') {
    const user = a.username ?? (placeholders ? 'username' : '')
    const pass = a.password ?? (placeholders ? 'password' : '')
    out.headers.push(['Authorization', `Basic ${btoa(`${user}:${pass}`)}`])
  } else if (s.type === 'apiKey') {
    const v = a.value || 'YOUR_API_KEY'
    const target = s.in === 'query' ? out.query : s.in === 'cookie' ? out.cookies : out.headers
    target.push([s.name || 'X-API-Key', v])
  } else if (s.type === 'http' && s.scheme && s.scheme !== 'bearer') {
    out.headers.push(['Authorization', `${s.scheme[0].toUpperCase()}${s.scheme.slice(1)} ${a.token || 'YOUR_TOKEN'}`])
  } else {
    // bearer, oauth2 and openIdConnect all end up as a bearer token.
    out.headers.push(['Authorization', `Bearer ${a.token || 'YOUR_SECRET_TOKEN'}`])
  }
  return out
}

export function buildRequest(op: ApiOperation, input: RequestInput, extra = applyAuthNone): BuiltRequest {
  const path = op.path.replace(/\{([^}]+)\}/g, (m, name: string) => {
    const v = input.path[name]
    return v && v !== m ? encodeURIComponent(v) : m
  })
  const query = [...input.query, ...extra.query].filter(([k]) => k)
  const qs = query.map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`).join('&')
  const headers: [string, string][] = [...extra.headers, ...input.headers.filter(([k]) => k)]
  const cookies = [...input.cookies, ...extra.cookies].filter(([k]) => k)
  if (cookies.length) headers.push(['Cookie', cookies.map(([k, v]) => `${k}=${v}`).join('; ')])
  const hasBody = input.body !== undefined && input.body !== '' && !['get', 'head'].includes(op.method)
  if (hasBody && input.contentType && !headers.some(([k]) => k.toLowerCase() === 'content-type')) {
    headers.push(['Content-Type', input.contentType])
  }
  return {
    method: op.method.toUpperCase(),
    url: (input.server || '') + path + (qs ? `?${qs}` : ''),
    headers,
    body: hasBody ? input.body : undefined,
  }
}

const applyAuthNone = { headers: [] as [string, string][], query: [] as [string, string][], cookies: [] as [string, string][] }

// ---- code samples --------------------------------------------------------

export type Client = { id: string; label: string; lang: string }

export const clients: Client[] = [
  { id: 'curl', label: 'cURL', lang: 'shell' },
  { id: 'js', label: 'JavaScript', lang: 'js' },
  { id: 'python', label: 'Python', lang: 'python' },
  { id: 'go', label: 'Go', lang: 'go' },
  { id: 'php', label: 'PHP', lang: 'php' },
  { id: 'http', label: 'HTTP', lang: 'http' },
]

const q1 = (s: string) => `'${s.replace(/'/g, `'\\''`)}'`
const dq = (s: string) => JSON.stringify(s)
const indent = (s: string, pad: string) => s.replace(/\n/g, `\n${pad}`)
const isJSON = (r: BuiltRequest) => r.headers.some(([k, v]) => k.toLowerCase() === 'content-type' && v.includes('json'))

export function snippet(client: string, r: BuiltRequest): string {
  const url = r.url || '/'
  switch (client) {
    case 'curl': {
      const lines = [`curl ${r.method === 'GET' ? '' : `--request ${r.method} \\\n  `}--url ${q1(url)}`]
      for (const [k, v] of r.headers) lines.push(`--header ${q1(`${k}: ${v}`)}`)
      if (r.body !== undefined) lines.push(`--data ${q1(r.body)}`)
      return lines.join(' \\\n  ')
    }
    case 'js': {
      const opts: string[] = []
      if (r.method !== 'GET') opts.push(`method: ${dq(r.method)}`)
      if (r.headers.length) {
        opts.push(`headers: {\n${r.headers.map(([k, v]) => `    ${dq(k)}: ${dq(v)}`).join(',\n')}\n  }`)
      }
      if (r.body !== undefined) {
        opts.push(isJSON(r) ? `body: JSON.stringify(${indent(r.body, '  ')})` : `body: ${dq(r.body)}`)
      }
      const args = opts.length ? `, {\n  ${opts.join(',\n  ')}\n}` : ''
      return `const response = await fetch(${dq(url)}${args})\n\nconst data = await response.json()\nconsole.log(data)`
    }
    case 'python': {
      const args = [dq(url)]
      if (r.headers.length) args.push(`headers={\n${r.headers.map(([k, v]) => `        ${dq(k)}: ${dq(v)}`).join(',\n')}\n    }`)
      if (r.body !== undefined) {
        args.push(isJSON(r) ? `json=${indent(pyLiteral(r.body), '    ')}` : `data=${dq(r.body)}`)
      }
      return `import requests\n\nresponse = requests.${r.method.toLowerCase()}(\n    ${args.join(',\n    ')}\n)\n\nprint(response.json())`
    }
    case 'go': {
      const lines = ['package main', '', 'import (', '\t"fmt"', '\t"io"', '\t"net/http"']
      if (r.body !== undefined) lines.push('\t"strings"')
      lines.push(')', '', 'func main() {')
      const body = r.body !== undefined ? `strings.NewReader(${goString(r.body)})` : 'nil'
      lines.push(`\treq, _ := http.NewRequest(${dq(r.method)}, ${dq(url)}, ${body})`)
      for (const [k, v] of r.headers) lines.push(`\treq.Header.Add(${dq(k)}, ${dq(v)})`)
      lines.push('', '\tres, err := http.DefaultClient.Do(req)', '\tif err != nil {', '\t\tpanic(err)', '\t}', '\tdefer res.Body.Close()', '', '\tbody, _ := io.ReadAll(res.Body)', '\tfmt.Println(string(body))', '}')
      return lines.join('\n')
    }
    case 'php': {
      const opts = [`CURLOPT_URL => ${q1(url)}`, 'CURLOPT_RETURNTRANSFER => true', `CURLOPT_CUSTOMREQUEST => '${r.method}'`]
      if (r.headers.length) opts.push(`CURLOPT_HTTPHEADER => [\n${r.headers.map(([k, v]) => `    ${q1(`${k}: ${v}`)}`).join(',\n')}\n  ]`)
      if (r.body !== undefined) opts.push(`CURLOPT_POSTFIELDS => ${q1(r.body)}`)
      return `<?php\n\n$curl = curl_init();\n\ncurl_setopt_array($curl, [\n  ${opts.join(',\n  ')},\n]);\n\n$response = curl_exec($curl);\ncurl_close($curl);\n\necho $response;`
    }
    case 'http': {
      let host = ''
      let target = url
      try {
        const u = new URL(url)
        host = u.host
        target = u.pathname + u.search
      } catch {
        // a relative server URL: no Host line
      }
      const lines = [`${r.method} ${target} HTTP/1.1`]
      if (host) lines.push(`Host: ${host}`)
      for (const [k, v] of r.headers) lines.push(`${k}: ${v}`)
      return lines.join('\n') + (r.body !== undefined ? `\n\n${r.body}` : '')
    }
  }
  return ''
}

function goString(s: string): string {
  return s.includes('`') ? dq(s) : `\`${s}\``
}

// pyLiteral turns JSON into a Python literal (true/false/null differ).
function pyLiteral(json: string): string {
  try {
    JSON.parse(json)
  } catch {
    return dq(json)
  }
  return json.replace(/("(?:[^"\\]|\\.)*")|\b(true|false|null)\b/g, (m, str: string | undefined, word: string | undefined) =>
    str ? m : word === 'true' ? 'True' : word === 'false' ? 'False' : 'None',
  )
}

// ---- highlighting ---------------------------------------------------------

// Token classes follow Chroma's, so the site's Chroma theme colours them.
type Rule = [RegExp, string]

const common: Rule[] = [
  [/^("(?:[^"\\\n]|\\.)*"|'(?:[^'\\\n]|\\.)*'|`[^`]*`)/, 's'],
  [/^-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b/, 'm'],
]

const kw = (words: string) => new RegExp(`^\\b(?:${words.split(' ').join('|')})\\b`)

const rules: Record<string, Rule[]> = {
  json: [
    [/^"(?:[^"\\\n]|\\.)*"(?=\s*:)/, 'nt'],
    [/^"(?:[^"\\\n]|\\.)*"/, 's'],
    [/^-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/, 'm'],
    [/^\b(?:true|false|null)\b/, 'kc'],
  ],
  shell: [[/^#.*/, 'c'], [/^(?:curl)\b/, 'nb'], [/^--?[a-zA-Z][\w-]*/, 'na'], ...common],
  js: [[/^\/\/.*/, 'c'], [kw('const let var await async function return new if else'), 'k'], [kw('true false null undefined'), 'kc'], [/^[A-Za-z_$][\w$]*(?=\()/, 'nf'], ...common],
  python: [[/^#.*/, 'c'], [kw('import from def return if else print'), 'k'], [kw('True False None'), 'kc'], [/^[A-Za-z_]\w*(?=\()/, 'nf'], ...common],
  go: [[/^\/\/.*/, 'c'], [kw('package import func if return defer var'), 'k'], [kw('nil true false'), 'kc'], [/^[A-Za-z_]\w*(?=\()/, 'nf'], ...common],
  php: [[/^\/\/.*/, 'c'], [/^<\?php/, 'k'], [/^\$\w+/, 'nv'], [kw('echo true false null'), 'k'], [/^[A-Z_]{3,}\b/, 'no'], [/^[A-Za-z_]\w*(?=\()/, 'nf'], ...common],
  http: [[/^(?:GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|TRACE)\b/, 'k'], [/^HTTP\/[\d.]+/, 'kc'], [/^[A-Za-z][\w-]*(?=:)/, 'na']],
}

const esc = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

export function highlight(code: string, lang: string): string {
  const rs = rules[lang] ?? (lang.includes('json') ? rules.json : [])
  if (!rs.length) return esc(code)
  let out = ''
  let rest = code
  let plain = ''
  while (rest) {
    let hit = false
    // Only start a token at a word boundary, so "a1" is not "a" + number.
    const prev = plain.slice(-1) || out.slice(-1)
    for (const [re, cls] of rs) {
      if (/[\w$]/.test(prev) && /^[\w$]/.test(rest)) break
      const m = re.exec(rest)
      if (m && m[0]) {
        out += esc(plain) + `<span class="${cls}">${esc(m[0])}</span>`
        plain = ''
        rest = rest.slice(m[0].length)
        hit = true
        break
      }
    }
    if (!hit) {
      plain += rest[0]
      rest = rest.slice(1)
    }
  }
  return out + esc(plain)
}

export function prettyJSON(v: unknown): string {
  if (v === undefined) return ''
  if (typeof v === 'string') return v
  return JSON.stringify(v, null, 2)
}

// statusTone groups HTTP status codes for colour: 2xx ok, 3xx, 4xx, 5xx.
export function statusTone(code: string | number): 'ok' | 'redirect' | 'client' | 'server' | 'other' {
  const c = String(code)[0]
  return c === '2' ? 'ok' : c === '3' ? 'redirect' : c === '4' ? 'client' : c === '5' ? 'server' : 'other'
}
