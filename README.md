# GoogleSQLite

GoogleSQLite is a Go project that runs GoogleSQL — the SQL dialect
used by BigQuery and Cloud Spanner — on top of a SQLite backend. It is
exposed as a `database/sql` driver, so any `database/sql` consumer gets
a local GoogleSQL execution engine with no external services to run.

GoogleSQL parsing and analysis is provided by
[`goccy/go-googlesql`](https://github.com/goccy/go-googlesql), and query
execution runs on
[`ncruces/go-sqlite3`](https://github.com/ncruces/go-sqlite3). Both are
implemented in pure Go, so GoogleSQLite is pure Go as well — no cgo is
required.

Because GoogleSQL is the dialect behind BigQuery and Cloud Spanner,
this project gives strong support to use cases such as locally
emulating those services.

## Usage

```go
import (
	"database/sql"

	_ "github.com/goccy/googlesqlite"
)

db, err := sql.Open("googlesqlite", ":memory:")
if err != nil {
	log.Fatal(err)
}
defer db.Close()

rows, err := db.Query(`SELECT FORMAT('%t', DATE '2026-05-14') AS today`)
// ...
```

The driver registers under `googlesqlite`. DSN is the underlying
SQLite path (`:memory:`, a file path, or
`file:foo.db?cache=shared`).

## Status

Every function and type is backed by a declarative spec under
`docs/specs/` and at least one upstream-sourced test case under
`testdata/specs/`. The full support matrix — per-function and
per-type status with a link to each spec — is generated from spec
frontmatter into [`docs/specs/INDEX.md`](docs/specs/INDEX.md).

Specs marked `partial` are documented gaps with a concrete
infrastructure dependency; see each spec's frontmatter `notes:`
field for what is missing.

## Quick start

```sh
make build            # build the bin/googlesqlite CLI
make test             # run unit and spec tests
make lint             # golangci-lint via tools/go.mod
make spec/upstream-sync   # refresh docs/third_party/googlesql-docs from upstream
```

## Layout

- `docs/specs/` — canonical, normalized spec markdown (one file per
  function/type).
- `testdata/specs/` — declarative YAML test cases referenced by spec
  frontmatter. The repo-root `TestSpec` runner executes each case
  against the driver as a Go subtest.
- `docs/third_party/googlesql-docs/` — Apache-2.0 snapshot of upstream
  `google/googlesql` `docs/`. Refreshed by
  `make spec/upstream-sync`.
- `cmd/specctl/` — operational CLI (normalize, upstream-sync,
  check, ...).
- `tools/go.mod` — pinned developer tools (golangci-lint).

## Sponsorship

This is a personal project. It is developed by referencing the
open-source GoogleSQL repository, but it receives no support of any kind
from Google — no sponsorship, no contributions, no promotion.

For example, Google has had a request for a BigQuery emulator open on
its Issue Tracker
([issue 129248927](https://issuetracker.google.com/issues/129248927))
for seven years without taking any action. Unable to watch that any
longer, in 2022 I built
[`goccy/bigquery-emulator`](https://github.com/goccy/bigquery-emulator)
— the only BigQuery emulator in the world — and have maintained it ever
since.

This project is used by that `bigquery-emulator`. I do not use BigQuery
in my own job, however. I simply happen to have the skills to build
this, noticed how many people are struggling without it, and so I spend
my personal time and money on it.

Because of that, keeping this project alive needs your help. Emulating
BigQuery locally brings many benefits to development and can
significantly cut development costs. Could you return a small part of
those savings to me? I believe doing so leads to a better future for
both me and your company.

See `CLAUDE.md` for the project's working rules.
