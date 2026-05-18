// pack-zoneinfo.mjs packs the host IANA timezone database into
// public/zoneinfo.json so the Playground worker can mount it in an
// in-memory filesystem for the wasm engine.
//
// The GoogleSQL wasm module reads timezone files from
// /usr/share/zoneinfo at start-up; a browser has no such filesystem,
// so the worker recreates it from this file.

import {
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const ZONEINFO_DIR = '/usr/share/zoneinfo'
const here = dirname(fileURLToPath(import.meta.url))
const outPath = join(here, '..', 'public', 'zoneinfo.json')

/** Recursively collect every file under dir, dereferencing symlinks. */
function walk(dir, files) {
  for (const name of readdirSync(dir)) {
    const full = join(dir, name)
    const st = statSync(full) // statSync follows symlinks
    if (st.isDirectory()) {
      walk(full, files)
    } else if (st.isFile()) {
      files[relative(ZONEINFO_DIR, full)] = readFileSync(full).toString('base64')
    }
  }
}

const files = {}
walk(ZONEINFO_DIR, files)

mkdirSync(dirname(outPath), { recursive: true })
writeFileSync(outPath, JSON.stringify(files))
console.log(
  `pack-zoneinfo: ${Object.keys(files).length} files -> public/zoneinfo.json`,
)
