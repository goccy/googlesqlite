# Project Rules

## Language

All code, comments, commit messages, PR descriptions, and documentation
files in this repository (including `testdata/**/*.yaml`,
`docs/specs/**/*.md`, generated reports, and CI scripts) MUST be written
in English. Conversation in chat may be in any language.

The plan files under `.claude/plans/` are an exception: they are working
notes maintained by request and may be in another language.

## Identity

The project's public identity is `googlesql` and `googlesqlite` only.

- Driver name: `googlesqlite` (single name, no aliases).
- Go module: `github.com/goccy/googlesqlite`.
- Public types live in `package googlesqlite` and are named `Driver`,
  `Conn`, `Tx`, etc.
- Do not introduce identifiers, file names, comments, or strings
  referring to predecessor branding. The only acceptable mention of the
  predecessor is in commit messages or PR descriptions when describing
  history.
- `make identity-check` enforces this: it fails the build if predecessor
  strings appear anywhere in tracked files outside `docs/third_party/`,
  `.claude/plans/`, and the Makefile rule itself.

## Tool dependencies

Every Go-based developer tool used by this project (linters, formatters,
code generators, etc.) MUST be managed through `tools/go.mod` using Go's
`tool` directive. The pinned versions in `tools/go.mod` and
`tools/go.sum` are authoritative — every contributor and every CI job
runs the exact same version.

Add a tool:

    go get -tool -modfile=tools/go.mod <package>@<version>

Run a tool:

    go tool -modfile=tools/go.mod <tool-name> [args...]

Use the Makefile aliases (`make lint`, `make lint/fix`, etc.) when one
exists; they already pass `-modfile=tools/go.mod`. Never invoke project
tools through ambient `$GOBIN`, `go install`, or the root `go.mod` —
that path makes contributors' results depend on whatever they happen to
have installed locally.

Currently managed tools:

- `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`
- `github.com/goreleaser/goreleaser/v2` (run via `make release`)

### Non-Go tools

Tools that are not Go programs cannot live in `tools/go.mod`. They MUST
still be pinned and obtained by a single mechanism shared between local
builds and CI — never an ad-hoc `brew install` locally versus a
different install path in a workflow.

Currently the only such tool is **Binaryen** (`wasm-opt`, used to
optimise the Playground engine). It is pinned and downloaded entirely
by `docs/playground/Makefile`: `BINARYEN_VERSION` pins the release and
the `$(WASMOPT)` file target downloads, checksum-verifies and unpacks
it under `tools/binaryen-<version>/` on demand. `make optimize` (and
therefore `make build/release` and `make release`) depends on that
target, so the release workflow needs no separate install step.

## Compatibility

This project provides a `database/sql` driver registered as
`googlesqlite`. It serves as the local execution backend for GoogleSQL
workloads and is intended to be embedded by downstream consumers.

The public surface — driver name, DSN format, and the
`Driver`/`Conn`/`Tx` interfaces along with the exported methods on
`Conn` (`SetNamePath`, `SetMaxNamePath`) — MUST stay stable so that
downstream consumers can depend on it without structural changes to
their own code.

Keep the public surface minimal: only export what an external consumer
actually needs. Functionality exercised solely by the in-repo test
suite belongs in `export_test.go` (re-exported for the black-box
`googlesqlite_test` package) rather than in the production API. The
wasm runtime is configured through the `EnvWasm*` environment-variable
constants, not a programmatic setter.

Performance MUST NOT regress. The `bench/` suite exercises
representative GoogleSQL workloads and is run before any release; it
measures the driver against a saved baseline of its own previous run
(`bench/RESULTS.md`) to catch regressions across revisions.
Window-function-heavy queries in particular must stay fast, because
this driver registers custom window aggregates natively rather than
falling back to SQLite's built-in window functions.

## Testdata is authoritative — never trim it to dodge a failure

This rule overrides any other progress incentive in this repository.

- **Testdata expected values MUST come from an authoritative source.**
  Acceptable sources: the upstream `docs/third_party/googlesql-docs/**`
  Examples sections, the GoogleSQL `compliance/testdata/*.test`
  fixtures, the Cloud SQL / BigQuery / Spanner reference docs, or the
  Cloud Spanner Emulator conformance suite. **Never** fill an
  `expected:` field from observed runtime output.
- **When the upstream source provides multiple Examples, every one of
  them MUST appear in the spec's testdata YAML and every one MUST
  pass.** Reducing a spec's case list because the runtime cannot
  satisfy a subset is forbidden. Replacing a specific upstream
  assertion (e.g. `expected: [["abc"]]`) with a weaker property
  (`expected: [[true]]` via `IS NOT NULL`) is the same violation.
- **Everything is implementable unless `go-googlesql` itself can't
  support it.** "I cannot implement this" is not an acceptable reason
  to weaken or remove testdata, or to stop work. The only legitimate
  way to make a failing case pass is to fix the implementation, the
  analyzer wiring, or the formatter — never the test.
- **Do not stop work until every test passes.** Mid-task stops because
  of perceived complexity are not allowed; pick the next failing case
  and fix it.
- Trimming existing testdata is considered one of the most serious
  rule violations in this repository. If you are tempted to do it,
  stop and fix the underlying implementation instead.

## Project layout (high-level)

```
.            top-level package googlesqlite (driver.go etc.) plus the
             black-box test suite (driver/query/ddl/dml/types/... _test.go)
cmd/         executables (googlesqlite CLI, specctl)
internal/    driver implementation: functions/, value/, sqlitex/,
             specmeta/, plus the analyzer/formatter glue
docs/specs/  canonical, normalized spec markdown (one per function/type)
testdata/    YAML test cases referenced by spec frontmatter
bench/       performance corpus and runner
docs/third_party/googlesql-docs/  Apache-2.0 snapshot of upstream googlesql/docs/
tools/       tool-only Go module (golangci-lint, etc.)
```

See `.claude/plans/` for the working architectural plan.
