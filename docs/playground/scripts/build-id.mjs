// build-id.mjs records a content hash and the uncompressed size of the
// engine binary in public/build-id.json. The worker appends the hash
// as a ?v= query so a new build busts the browser cache, and uses the
// size as the loading-progress total (accurate even when the binary is
// served gzip-compressed).

import { createHash } from 'node:crypto'
import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const wasmPath = join(here, '..', 'public', 'googlesqlite.wasm')
const outPath = join(here, '..', 'public', 'build-id.json')

const bytes = readFileSync(wasmPath)
const hash = createHash('sha256').update(bytes).digest('hex').slice(0, 16)
writeFileSync(outPath, JSON.stringify({ wasm: hash, wasmSize: bytes.length }))
console.log(`build-id: wasm ${hash} (${bytes.length} bytes)`)
