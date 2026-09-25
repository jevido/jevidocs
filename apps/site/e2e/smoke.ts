// Browser smoke tests for the docs site, the admin and the API.
//
//   bun run e2e                          # against local dev servers
//   SITE_URL=https://jevidocs.jevido.app ADMIN_URL=https://admin.jevidocs.jevido.app \
//     API_URL=https://api.jevidocs.jevido.app bun run e2e
//
// Admin checks run only when ADMIN_EMAIL and ADMIN_PASSWORD are set. Any
// console error or uncaught page error fails the run.
import puppeteer, { type Browser, type Page } from 'puppeteer-core'

const SITE = (process.env.SITE_URL ?? 'http://127.0.0.1:4720').replace(/\/$/, '')
const ADMIN = (process.env.ADMIN_URL ?? 'http://127.0.0.1:4740').replace(/\/$/, '')
const API = (process.env.API_URL ?? 'http://127.0.0.1:4730').replace(/\/$/, '')
const CHROME = process.env.CHROME_PATH ?? '/usr/bin/chromium'
const EMAIL = process.env.ADMIN_EMAIL
const PASSWORD = process.env.ADMIN_PASSWORD

type Result = { name: string; status: 'pass' | 'fail' | 'skip'; detail?: string }
const results: Result[] = []

class Skip extends Error {}

function assert(cond: unknown, msg: string): asserts cond {
  if (!cond) throw new Error(msg)
}

async function test(name: string, fn: () => Promise<void>) {
  const started = performance.now()
  try {
    await fn()
    results.push({ name, status: 'pass', detail: `${Math.round(performance.now() - started)}ms` })
  } catch (e) {
    if (e instanceof Skip) results.push({ name, status: 'skip', detail: e.message })
    else results.push({ name, status: 'fail', detail: e instanceof Error ? e.message : String(e) })
  }
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

async function waitFor(fn: () => Promise<boolean>, what: string, timeout = 10000) {
  const until = Date.now() + timeout
  while (Date.now() < until) {
    if (await fn()) return
    await sleep(100)
  }
  throw new Error(`timed out waiting for ${what}`)
}

async function json(path: string, init?: RequestInit) {
  const res = await fetch(API + path, init)
  assert(res.ok, `${path}: HTTP ${res.status}`)
  return res.json()
}

// --- API ----------------------------------------------------------------------

await test('api: health', async () => {
  const h = await json('/health')
  assert(h.status === 'ok' && h.database === 'ok', `health = ${JSON.stringify(h)}`)
})

await test('api: search finds tabs', async () => {
  const r = await json('/api/projects/jevidocs/search?q=tabs')
  assert(Array.isArray(r) && r.some((x: { slug: string }) => x.slug === 'components/tabs'), 'no components/tabs result')
})

await test('api: llms.txt', async () => {
  const res = await fetch(API + '/api/projects/jevidocs/llms.txt')
  const text = await res.text()
  assert(res.ok && text.startsWith('# jevidocs') && text.includes('page.md?slug='), 'unexpected llms.txt')
})

await test('api: MCP tools/list', async () => {
  const r = await json('/mcp', {
    method: 'POST',
    headers: { 'content-type': 'application/json', accept: 'application/json, text/event-stream' },
    body: JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'tools/list' }),
  })
  const names: string[] = r.result.tools.map((t: { name: string }) => t.name)
  for (const n of ['search_docs', 'read_page', 'create_page']) assert(names.includes(n), `missing tool ${n}`)
})

// --- Browser ------------------------------------------------------------------

const browser: Browser = await puppeteer.launch({ executablePath: CHROME, args: ['--no-sandbox'] })
const consoleErrors: string[] = []

async function newPage(width = 1400, height = 900): Promise<Page> {
  const page = await browser.newPage()
  await page.setViewport({ width, height })
  page.on('pageerror', (e) => consoleErrors.push(`${page.url()}: ${e}`))
  page.on('console', (m) => {
    if (m.type() !== 'error') return
    const text = m.text()
    // An intentional 404 (the not-found test) logs a failed resource load.
    if (page.url().includes('/definitely-not-a-page') && text.includes('404')) return
    consoleErrors.push(`${page.url()}: ${text}`)
  })
  return page
}

async function goto(page: Page, url: string) {
  await page.goto(url, { waitUntil: 'networkidle0', timeout: 30000 })
}

async function waitTitle(page: Page, text: string) {
  await waitFor(async () => (await page.$eval('h1.title', (e) => e.textContent ?? '').catch(() => '')).includes(text), `h1 "${text}"`)
}

const page = await newPage()

await test('site: landing page', async () => {
  await goto(page, SITE + '/')
  const h1 = await page.$eval('h1', (e) => e.textContent ?? '')
  assert(h1.toLowerCase().includes('documentation'), `h1 = ${h1}`)
  assert(await page.$('a[href="/docs"], a[href^="/docs"]'), 'no link to /docs')
})

await test('site: docs page with sidebar, TOC and breadcrumbs', async () => {
  await goto(page, SITE + '/docs/components/tabs')
  await waitTitle(page, 'Tabs')
  assert((await page.$$('.sidebar a.item')).length > 5, 'sidebar has too few links')
  assert((await page.$$('nav.toc a')).length > 0, 'TOC is empty')
  const crumbs = await page.$eval('nav.crumbs', (e) => e.textContent ?? '')
  assert(crumbs.includes('Components'), `breadcrumbs = ${crumbs}`)
})

await test('site: tabs switch', async () => {
  const triggers = await page.$$('.fd-tab-trigger')
  assert(triggers.length >= 2, 'no tabs on the page')
  const value = await triggers[1].evaluate((e) => e.getAttribute('data-tab'))
  await triggers[1].click()
  const active = await page.$eval(`.fd-tab[data-value="${value}"]`, (e) => e.hasAttribute('data-active'))
  assert(active, 'second tab did not become active')
})

await test('site: SPA navigation keeps state', async () => {
  await page.evaluate(() => ((window as unknown as { __e2e: number }).__e2e = 1))
  await page.click('.sidebar a.item[href="/docs/quick-start"]')
  await waitTitle(page, 'Quick start')
  const kept = await page.evaluate(() => (window as unknown as { __e2e?: number }).__e2e)
  assert(kept === 1, 'page reloaded instead of client-side navigation')
  assert(page.url().endsWith('/docs/quick-start'), `url = ${page.url()}`)
})

await test('site: search dialog (Ctrl+K, Enter)', async () => {
  await page.keyboard.down('Control')
  await page.keyboard.press('k')
  await page.keyboard.up('Control')
  await page.waitForSelector('[role="dialog"] input', { timeout: 5000 })
  await page.keyboard.type('callout')
  await waitFor(async () => (await page.$$('.result')).length > 0, 'search results')
  await page.keyboard.press('Enter')
  await waitFor(async () => page.url().includes('/docs/components/callout'), 'navigation to callout')
  await waitTitle(page, 'Callout')
})

await test('site: root toggle lists REST API', async () => {
  await goto(page, SITE + '/docs/api')
  await page.waitForSelector('.root-toggle .trigger', { timeout: 10000 })
  await page.click('.root-toggle .trigger')
  const items = await page.$$eval('.root-toggle .menu a', (es) => es.map((e) => e.textContent ?? ''))
  assert(items.some((t) => t.includes('REST API')), `menu = ${items.join(' | ')}`)
})

await test('site: OpenAPI reference with outline and request client', async () => {
  await goto(page, SITE + '/docs/api/reference#tag/pages/GET/api/projects/{project}/search')
  await waitFor(async () => (await page.$$('section.op[data-op]')).length >= 7, 'operations')
  assert((await page.$$('.sidebar .api-nav a.op')).length >= 7, 'sidebar outline is missing operations')
  assert((await page.$$('nav.toc a')).length === 0, 'an API reference has no TOC column')
  const code = await page.$eval('section.op .panel pre', (e) => e.textContent ?? '')
  assert(code.includes('curl'), `request sample = ${code.slice(0, 80)}`)
  await page.click('section.op button.test')
  await waitFor(async () => (await page.$('dialog.client[open]')) !== null, 'request client')
  await page.keyboard.press('Escape')
})

await test('site: mermaid renders on a canvas', async () => {
  const cdn = await fetch('https://cdn.jsdelivr.net/npm/mermaid@11/package.json').catch(() => null)
  if (!cdn?.ok) throw new Skip('jsdelivr unreachable')
  await goto(page, SITE + '/docs/writing/markdown')
  await waitFor(async () => (await page.$$('.fd-mermaid .fd-canvas svg')).length > 0, 'mermaid canvas', 20000)
})

await test('site: theme toggle', async () => {
  await page.click('.toggle button[aria-label="Dark"]')
  assert(await page.evaluate(() => document.documentElement.classList.contains('dark')), 'dark class not set')
  await page.click('.toggle button[aria-label="Light"]')
  assert(!(await page.evaluate(() => document.documentElement.classList.contains('dark'))), 'dark class not removed')
  await page.evaluate(() => {
    try {
      localStorage.removeItem('jevidocs-theme')
    } catch {
      // ignore
    }
  })
})

await test('site: not-found state', async () => {
  await goto(page, SITE + '/docs/definitely-not-a-page')
  await waitFor(async () => (await page.title()).startsWith('Not found'), 'not-found title')
  const code = await page.$eval('p.code', (e) => e.textContent ?? '')
  assert(code.includes('404'), 'no 404 marker')
})

await test('site: mobile drawer', async () => {
  const m = await newPage(400, 800)
  await goto(m, SITE + '/docs')
  await m.waitForSelector('button.menu', { visible: true, timeout: 10000 })
  await m.click('button.menu')
  await waitFor(async () => (await m.$('.sidebar.open')) !== null, 'open drawer')
  await m.close()
})

await test('site: project directory', async () => {
  await goto(page, SITE + '/p/')
  await waitFor(async () => (await page.$$('.card .name')).length > 0, 'project cards')
  const names = await page.$$eval('.card .name', (es) => es.map((e) => e.textContent))
  assert(names.includes('jevidocs'), `projects = ${names.join(', ')}`)
})

// --- Admin --------------------------------------------------------------------

await test('admin: login, dashboard, project, editor preview', async () => {
  if (!EMAIL || !PASSWORD) throw new Skip('ADMIN_EMAIL / ADMIN_PASSWORD not set')
  const a = await newPage()
  await goto(a, ADMIN + '/')
  await a.type('input[type=email]', EMAIL)
  await a.type('input[type=password]', PASSWORD)
  await a.click('button[type=submit]')
  await waitFor(async () => (await a.$('input[type=password]')) === null, 'leaving the login screen')
  await goto(a, ADMIN + '/#/projects/jevidocs')
  await waitFor(async () => (await a.$$('table tbody tr')).length > 5, 'page list')
  const token = await a.evaluate(() => {
    const key = Object.keys(localStorage).find((k) => k.includes('token'))
    return key ? localStorage.getItem(key) : null
  })
  assert(token, 'no session token stored')
  const pages = await json('/api/admin/projects/jevidocs/pages', { headers: { authorization: `Bearer ${token}` } })
  const tabs = pages.find((p: { slug: string }) => p.slug === 'components/tabs')
  assert(tabs, 'components/tabs page missing')
  await goto(a, `${ADMIN}/#/projects/jevidocs/pages/${tabs.id}`)
  await waitFor(async () => (await a.$$('.fd-prose .fd-tabs')).length > 0, 'rendered preview', 15000)
  await a.close()
})

await browser.close()

await test('no console errors', async () => {
  assert(consoleErrors.length === 0, consoleErrors.join('\n'))
})

// --- Report -------------------------------------------------------------------

const icon = { pass: '✓', fail: '✗', skip: '-' }
console.log(`\nsite ${SITE}  admin ${ADMIN}  api ${API}\n`)
for (const r of results) console.log(`${icon[r.status]} ${r.name}${r.detail ? `  (${r.detail})` : ''}`)
const failed = results.filter((r) => r.status === 'fail').length
console.log(`\n${results.length - failed} ok, ${failed} failed`)
process.exit(failed ? 1 : 0)
