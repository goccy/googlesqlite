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
$ make build         # builds the wasm engine and bundles the site into dist/
$ make build/release # same, but shrinks the engine with wasm-opt first
$ make preview       # serve the production build locally
```

`make build/release` is what the release workflow runs: it optimises
the engine with `wasm-opt` before bundling, so it additionally needs
Binaryen installed (see *Optimising the wasm binary* below).

## How it works

`make build/wasm` compiles `cmd/googlesqlite` (the `js && wasm` build)
into `public/googlesqlite.wasm` and copies the Go `wasm_exec.js` loader
next to it. Both files are generated artifacts and are gitignored.

The worker (`src/api/worker.ts`) loads `wasm_exec.js`, instantiates the
engine, and exposes `exec` / history / export calls. The UI thread
talks to it through `src/api/googlesqlite.ts`.

## Optimising the wasm binary

The engine binary is large: it bundles the whole googlesqlite stack
into one js/wasm module — SQLite (transpiled to Go) plus the GoogleSQL
analyzer, which is itself an embedded wasm module run through wazero.
`wasm-opt` (Binaryen) can shrink the code section and speed up
start-up:

```console
$ make optimize   # runs wasm-opt over public/googlesqlite.wasm
```

`make optimize` needs no manual setup: the Makefile downloads and
checksum-verifies a pinned Binaryen release (`BINARYEN_VERSION`) under
`../../tools/binaryen-<version>/` on demand. Local builds and the
release workflow share this one mechanism, so they optimise with the
identical `wasm-opt`. `make build/release` runs this step for you.

Most of the binary is the embedded GoogleSQL analyzer wasm module,
which `wasm-opt` cannot compress further, so expect only a modest
reduction.

## Deployment

The site is deployed to GitHub Pages by the `Release` workflow
(`.github/workflows/release.yml`) whenever a `v*` tag is pushed. The
workflow runs `make build/release`, uploads `dist/` as the Pages
artifact, and publishes it through `actions/deploy-pages`.

The uncompressed engine binary exceeds GitHub's 100 MB per-file limit,
but it is never deployed: the worker fetches the gzip-compressed
`googlesqlite.wasm.gz` (tens of MB) and decompresses it in the browser,
and `make build/page` drops the uncompressed copy from `dist/`. The
same release workflow also attaches the optimised `googlesqlite.wasm`
to the GitHub release as a downloadable asset.
