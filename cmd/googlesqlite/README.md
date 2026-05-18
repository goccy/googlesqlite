# googlesqlite CLI

`googlesqlite` is an interactive console for running GoogleSQL queries
against the SQLite-backed googlesqlite engine, in the spirit of the
`sqlite3` and `spanner-cli` REPLs.

The same program also builds for `js/wasm` and powers the browser
Playground.

## Install

```console
$ go install github.com/goccy/googlesqlite/cmd/googlesqlite@latest
```

Or build from a checkout:

```console
$ make build        # native binary -> bin/googlesqlite
$ make build-wasm   # bin/googlesqlite.wasm + bin/wasm_exec.js
```

## Usage

```
googlesqlite [flags] [database]

Flags:
  --dsn string            data source name (overrides the positional database argument)
  --file value            run statements from a SQL file before the REPL starts (repeatable)
  --debug                 show the translated SQLite query for each statement
  --no-color              disable coloured output
  --history string        REPL history file (default ~/.googlesqlite_history)
  --continue-on-error     keep running a script after a statement fails
```

With no `database` argument the CLI uses a shared in-memory database.
A positional argument opens (or creates) that database file.

### Input modes

All four input paths run the same statement pipeline; whenever a
terminal is available the REPL opens afterwards in the resulting state.

- **Interactive** — type statements at the prompt. Input is accumulated
  across lines until a `;` (or a trailing `\G`).
- **Startup file** — `--file script.sql` runs the file, then the REPL
  opens. `--file` may be repeated.
- **Piped stdin** — `cat script.sql | googlesqlite` runs the piped SQL,
  then the REPL opens on the controlling terminal (`/dev/tty`).
- **`.read`** — source a file mid-session.

## Dot-commands

- `.help` — show help
- `.quit`, `.exit` — leave the CLI
- `.debug [on|off]` — show the translated SQLite query for each statement
- `.tables` — list tables and views
- `.functions` — list functions
- `.read <path>` — run the statements in a SQL file

## Display modes

Results print as an aligned table. Column widths are measured by
display width, so full-width (CJK) text stays aligned. Append `\G` to a
query to print it vertically, one stanza per row.

## WebAssembly Playground API

The `js/wasm` build installs `globalThis.googlesqlite` with:

| function | description |
|----------|-------------|
| `exec(sql)` / `execScript(text)` | run statements; returns `{ output, results }` |
| `exportDB()` | return the database as a `Uint8Array` |
| `importDB(bytes)` | replace the session database |
| `getHistory()` | return the query/result history |
| `loadHistory(json)` | restore a history exported as JSON |
| `clearHistory()` | drop all history |
| `exportHistory(format)` | export history as `json`, `sql`, `markdown` or `csv` |
| `setDebug(bool)` | toggle the translated-SQLite display |

The host page is responsible for persisting `exportDB()` bytes and the
history (e.g. in IndexedDB) and restoring them on load via `importDB`
and `loadHistory`. Load the module with the Go `wasm_exec.js` shipped
alongside `googlesqlite.wasm`; call code after `onGoogleSQLiteReady`
fires.
