# googlesqlite Playground

A browser-based playground that runs GoogleSQL queries entirely
client-side, powered by `googlesqlite` compiled to WebAssembly.

- **Vite + React + TypeScript + MUI** for the UI
- **CodeMirror 6** as the SQL editor
- The `googlesqlite` engine runs in a **Web Worker** so the UI stays
  responsive
- The database is a single file in the **Origin Private File System**
  (OPFS); SQLite reads and writes it directly, so it survives reloads
  with no explicit save step

## Prerequisites

- Go (the version pinned in the repository `go.mod`)
- Node.js 20+ and npm

## Run locally

```console
$ cd docs/playground
$ make deps      # install npm dependencies (first time only)
$ make dev       # build the wasm engine, then start the dev server
```

`make dev` opens the page at <http://localhost:5173/googlesqlite/>.

## Build the static site

```console
$ make build     # builds the engine and bundles the site into dist/
$ make preview   # serve the production build locally
```

## How it works

`make build/wasm` compiles `cmd/googlesqlite` (the `js && wasm` build)
into `public/googlesqlite.wasm` and copies the Go `wasm_exec.js` loader
next to it. Both files are generated artifacts and are gitignored.

The worker (`src/api/worker.ts`) loads `wasm_exec.js`, instantiates the
engine, and exposes `exec` / history / export calls. The UI thread
talks to it through `src/api/googlesqlite.ts`.

## Deployment

The site is deployed to GitHub Pages by the `Release` workflow
(`.github/workflows/release.yml`) whenever a `v*` tag is pushed. The
workflow builds the engine, bundles the site, uploads `dist/` as the
Pages artifact, and publishes it through `actions/deploy-pages`.

The engine binary is large — it bundles the whole googlesqlite stack
into one js/wasm module: SQLite (transpiled to Go) plus the GoogleSQL
analyzer, which is itself an embedded wasm module run through wazero.
The uncompressed binary exceeds GitHub's 100 MB per-file limit, but it
is never deployed: the worker fetches the gzip-compressed
`googlesqlite.wasm.gz` (tens of MB) and decompresses it in the browser,
and `make build/page` drops the uncompressed copy from `dist/`. The
same release workflow also attaches `googlesqlite.wasm` to the GitHub
release as a downloadable asset.
