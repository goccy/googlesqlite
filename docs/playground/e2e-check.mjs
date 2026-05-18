import { chromium, firefox, webkit } from 'playwright'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const URL = process.env.PLAYGROUND_URL || 'http://localhost:5173/googlesqlite/'
const ENGINES = { chromium, firefox, webkit }
const engineName = process.env.BROWSER || 'chromium'
const engine = ENGINES[engineName]
if (!engine) {
  throw new Error(`unknown BROWSER: ${engineName}`)
}
console.log(`browser: ${engineName}`)

// A persistent context (a real on-disk profile) is required: WebKit's
// OPFS — which the Playground uses to persist the database — fails in
// an ephemeral context. A fresh directory per run keeps runs isolated.
const userDataDir = mkdtempSync(join(tmpdir(), 'googlesqlite-e2e-'))
const context = await engine.launchPersistentContext(userDataDir, {
  viewport: { width: 1280, height: 800 },
})
const page = context.pages()[0] ?? (await context.newPage())

const logs = []
const wasmRequests = []
page.on('console', (m) => logs.push(`[console.${m.type()}] ${m.text()}`))
page.on('pageerror', (e) => logs.push(`[pageerror] ${e.message}`))
page.on('request', (r) => {
  if (r.url().includes('googlesqlite.wasm')) {
    wasmRequests.push(r.url())
  }
})

const fail = async (msg) => {
  console.log('CHECK FAILED:', msg)
  try {
    const body = await page.evaluate(() => document.body.innerText)
    console.log('=== page text ===')
    console.log(body.slice(0, 1000))
  } catch {
    // ignore
  }
  console.log('=== console / errors ===')
  console.log(logs.join('\n') || '(none)')
  await context.close()
  rmSync(userDataDir, { recursive: true, force: true })
  process.exit(1)
}

async function waitReady(p) {
  try {
    await p.waitForFunction(
      () => {
        const t = document.body.innerText
        return (
          t.includes('never uploaded to a server') ||
          t.includes('Failed to start')
        )
      },
      { timeout: 180000 },
    )
  } catch {
    await fail('engine never became ready (timeout)')
  }
  if ((await p.evaluate(() => document.body.innerText)).includes('Failed to start')) {
    await fail('engine reported an init error')
  }
}

// Each of these self-contained predicates is serialised and evaluated
// in the page: the Run button is disabled while a query executes.
const RUN_BUSY = () =>
  [...document.querySelectorAll('button')].some(
    (b) => /Run/.test(b.textContent || '') && b.disabled,
  )
const RUN_IDLE = () =>
  [...document.querySelectorAll('button')].some(
    (b) => /Run/.test(b.textContent || '') && !b.disabled,
  )

async function runQuery(p, sql) {
  await p.locator('.cm-content').click()
  await p.keyboard.press(process.platform === 'darwin' ? 'Meta+A' : 'Control+A')
  await p.keyboard.press('Delete')
  await p.keyboard.insertText(sql)
  await p.getByRole('button', { name: /^Run/i }).click()
  // Wait for the run to start (button disabled), then finish. The
  // start wait tolerates a query too fast to observe.
  await p.waitForFunction(RUN_BUSY, undefined, { timeout: 10000 }).catch(() => {})
  await p.waitForFunction(RUN_IDLE, undefined, { timeout: 60000 })
}

console.log('navigating to', URL)
const t0 = Date.now()
await page.goto(URL, { waitUntil: 'load', timeout: 60000 })

// Capture the loading screen (the engine is still initialising here).
await page.screenshot({ path: 'e2e-loading.png' })
const loadingText = await page.evaluate(() => document.body.innerText)
console.log(
  'loading screen shows progress:',
  /Downloading the engine|Starting the engine|Loading timezone/.test(loadingText),
)

await waitReady(page)
console.log(`engine ready (first load, ${Date.now() - t0} ms)`)

// The engine binary is fetched as a versioned, gzip-compressed file.
const first = wasmRequests[0]
if (!first || !first.includes('googlesqlite.wasm.gz') || !first.includes('?v=')) {
  await fail(`engine binary not fetched as a versioned .gz (${first})`)
}
console.log('engine binary fetched as a versioned .gz')

// OPFS persists the database across reloads — and so across e2e runs
// against the same origin. Drop any tables a previous run left behind
// so this run starts from a clean, predictable state.
await runQuery(
  page,
  'DROP TABLE IF EXISTS t;\n' +
    'DROP TABLE IF EXISTS users;\n' +
    'DROP TABLE IF EXISTS persist_test;',
)
console.log('clean start OK')

// Timezone query — the original failure mode.
await runQuery(
  page,
  "SELECT FORMAT_TIMESTAMP('%Y-%m-%d %H:%M', TIMESTAMP '2024-06-01 00:00:00+00', 'Asia/Tokyo') AS tokyo;",
)
if (!(await page.evaluate(() => document.body.innerText)).includes('2024-06-01 09:00')) {
  await fail('timezone query did not return the expected value')
}
console.log('timezone query OK')

// DDL + DML + SELECT with Japanese text.
await runQuery(
  page,
  'CREATE TABLE t (id INT64, name STRING);\n' +
    "INSERT INTO t VALUES (1, '山田'), (2, 'Bob');\n" +
    'SELECT id, name FROM t ORDER BY id;',
)
const deskText = await page.evaluate(() => document.body.innerText)
if (!deskText.includes('山田') || !deskText.includes('Bob')) {
  await fail('DDL/DML/SELECT round-trip did not show the inserted rows')
}
console.log('DDL/DML/SELECT OK (desktop)')

// Tapping a history entry restores that query and its result. Run a
// long-lined query (it must wrap in the editor), then another, then
// tap the first entry in the history.
const longQuery =
  "SELECT 'HISTORY_MARKER' AS tag_a, 1 AS a, 2 AS b, 3 AS c, 4 AS d, " +
  '5 AS e, 6 AS f, 7 AS g, 8 AS h, 9 AS i, 10 AS j;'
await runQuery(page, longQuery)
await runQuery(page, "SELECT 'second-query' AS col_b;")
await page.getByText("SELECT 'HISTORY_MARKER' AS tag_a").first().click()

// Tapping must restore both the query text (editor) and the result
// table for that entry.
const restored = await page
  .waitForFunction(
    () => {
      const editor = document.querySelector('.cm-content')
      const inEditor = !!editor && /HISTORY_MARKER/.test(editor.textContent || '')
      const inTable = [...document.querySelectorAll('table')].some((t) =>
        /HISTORY_MARKER/.test(t.textContent || ''),
      )
      return inEditor && inTable
    },
    undefined,
    { timeout: 10000 },
  )
  .then(() => true)
  .catch(() => false)
if (!restored) {
  const editorText = await page.locator('.cm-content').innerText()
  await fail(`history tap did not restore query + result (editor=${editorText})`)
}
console.log('history tap restores query + result OK')

// Privacy notice is shown.
if (
  !(await page.evaluate(() => document.body.innerText)).includes(
    'never uploaded to a server',
  )
) {
  await fail('privacy notice is not shown')
}
console.log('privacy notice shown OK')

// Examples menu loads an example into the editor without running it.
await page.getByRole('button', { name: /Examples/ }).click()
await page.getByRole('menuitem', { name: 'Create a table' }).click()
await page.waitForFunction(
  () => /CREATE TABLE users/.test(document.querySelector('.cm-content')?.textContent || ''),
  undefined,
  { timeout: 5000 },
)
console.log('examples menu loads the editor OK')
// Run it so the table-panel check below has a table to show.
await page.getByRole('button', { name: /^Run/i }).click()
await page.waitForFunction(RUN_BUSY, undefined, { timeout: 10000 }).catch(() => {})
await page.waitForFunction(RUN_IDLE, undefined, { timeout: 60000 })
if (!(await page.evaluate(() => document.body.innerText)).includes('Alice')) {
  await fail('running the loaded example did not produce its result')
}
console.log('examples menu OK')

// The Tables panel lists the created table; selecting it queries it.
await page
  .getByText('users', { exact: true })
  .first()
  .waitFor({ timeout: 10000 })
await page.getByText('users', { exact: true }).first().click()
await page.waitForFunction(RUN_BUSY, undefined, { timeout: 10000 }).catch(() => {})
await page.waitForFunction(RUN_IDLE, undefined, { timeout: 60000 })
const tablesEditor = await page.locator('.cm-content').innerText()
if (!tablesEditor.includes('SELECT * FROM')) {
  await fail('selecting a table did not query it')
}
console.log('tables panel OK')

// The Tables panel's trash button drops a table (its confirm dialog is
// auto-accepted). After the drop the table — and its button — are gone.
page.on('dialog', (dialog) => dialog.accept())
await page.getByRole('button', { name: /^drop users$/i }).click()
await page.waitForFunction(RUN_BUSY, undefined, { timeout: 10000 }).catch(() => {})
await page.waitForFunction(RUN_IDLE, undefined, { timeout: 60000 })
if ((await page.getByRole('button', { name: /^drop users$/i }).count()) !== 0) {
  await fail('dropping a table did not remove it from the Tables panel')
}
console.log('drop table from panel OK')

// Uploading a .sql file runs it.
await page.setInputFiles('input[type=file]', {
  name: 'uploaded.sql',
  mimeType: 'text/plain',
  buffer: Buffer.from("SELECT 'uploaded-file' AS source;"),
})
await page.waitForFunction(RUN_BUSY, undefined, { timeout: 10000 }).catch(() => {})
await page.waitForFunction(RUN_IDLE, undefined, { timeout: 60000 })
if (!(await page.evaluate(() => document.body.innerText)).includes('uploaded-file')) {
  await fail('uploading a .sql file did not run it')
}
console.log('file upload OK')

// Dark theme toggle changes the palette.
const bgLight = await page.evaluate(
  () => getComputedStyle(document.body).backgroundColor,
)
await page.getByRole('button', { name: /toggle theme/i }).click()
await page.waitForTimeout(300)
const bgDark = await page.evaluate(
  () => getComputedStyle(document.body).backgroundColor,
)
if (bgLight === bgDark) {
  await fail('theme toggle did not change the background colour')
}
await page.screenshot({ path: 'e2e-dark.png', fullPage: true })
console.log('dark theme toggle OK')
// Back to light for the remaining screenshot.
await page.getByRole('button', { name: /toggle theme/i }).click()
await page.waitForTimeout(200)

await page.waitForTimeout(400)
await page.screenshot({ path: 'e2e-desktop.png', fullPage: true })

// Create a sizable table so the reload below exercises that a real,
// non-trivial database survives in the OPFS-backed file.
await runQuery(
  page,
  'CREATE TABLE persist_test (id INT64, payload STRING);\n' +
    "INSERT INTO persist_test SELECT n, FORMAT('row-%d-aaaaaaaaaaaaaaaaaaaaaaaaaaaa', n) " +
    'FROM UNNEST(GENERATE_ARRAY(1, 5000)) AS n;',
)

// Reload: with the immutable Cache-Control the engine binary is served
// from the browser cache instead of re-downloading.
await page.setViewportSize({ width: 390, height: 844 })
const reloadStart = Date.now()
const wasmRequestsBefore = wasmRequests.length
await page.reload({ waitUntil: 'load' })
await waitReady(page)
console.log(`engine ready (reload, ${Date.now() - reloadStart} ms)`)

// The database must survive the reload: SQLite reads it straight from
// the OPFS file, so a new page sees the rows the old page wrote.
await runQuery(page, 'SELECT COUNT(*) AS n FROM persist_test;')
if (!(await page.evaluate(() => document.body.innerText)).includes('5000')) {
  await fail('table data did not survive a reload (persistence failed)')
}
console.log('database persisted across reload OK')

// On reload the engine binary must come from Cache Storage — but
// Cache Storage exists only in a secure context. Over plain HTTP it is
// absent, and a reload re-fetches; that must not be an error.
const reloadWasm = wasmRequests.slice(wasmRequestsBefore)
const secureContext = await page.evaluate(() => self.isSecureContext)
if (!secureContext) {
  console.log(
    `non-secure context: Cache Storage unavailable; reload re-fetched ` +
      `(${reloadWasm.length}) — engine still started`,
  )
} else if (reloadWasm.length === 0) {
  console.log('reload served the engine from Cache Storage (no re-download)')
} else {
  // Cache Storage persistence across reloads varies by browser and by
  // profile type; hard-fail only for the reference engine.
  const msg = `engine re-fetched on reload despite a secure context`
  if (engineName === 'chromium') {
    await fail(msg)
  }
  console.log(`WARNING (${engineName}): ${msg}`)
}
await page.waitForTimeout(400)
await page.screenshot({ path: 'e2e-mobile.png', fullPage: true })

// The history drawer opens from the app bar on mobile.
await page.getByRole('button', { name: 'history' }).click()
await page.waitForTimeout(600)
await page.screenshot({ path: 'e2e-mobile-history.png' })
const mobileText = await page.evaluate(() => document.body.innerText)
if (!mobileText.includes('History')) {
  await fail('history drawer did not open on mobile')
}
console.log('mobile layout + history drawer OK')

console.log('=== console / errors ===')
console.log(logs.join('\n') || '(none)')
console.log('ALL CHECKS PASSED')
await context.close()
rmSync(userDataDir, { recursive: true, force: true })
