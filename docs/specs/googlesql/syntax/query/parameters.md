---
name: parameters
dialect: googlesql
category: syntax/query
status: implemented
source_url: https://cloud.google.com/bigquery/docs/parameterized-queries
upstream_url: https://cloud.google.com/bigquery/docs/parameterized-queries
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/syntax/query/parameters.yaml
---

# Query parameters

## Summary

The driver accepts parameterised queries through both positional `?`
placeholders and named `@name` references, and tolerates the two
forms appearing together in a single statement.

## Signatures

- Positional: `WHERE x = ? AND y = ?`
- Named: `WHERE x = @a AND y = @b`
- Mixed: `WHERE x = @a AND y = ?`

## Behavior

- Each `?` in the SQL consumes one positional value from the args
  slice in order of appearance.
- Each `@name` reference binds to the `sql.Named("name", value)` arg
  that carries the matching name (case-insensitive).
- Mixed mode: the driver rewrites every `?` to a synthetic
  `@gsl_pos_N` (N is the 1-based occurrence count) before handing the
  SQL to the analyzer. Any positional driver.NamedValue (Name == "")
  is re-bound to the corresponding synthetic name in declaration
  order. This makes the parameter set effectively named, which the
  analyzer accepts even though its `ParameterMode` enum is a single
  flag.
- The rewrite is quote-aware: `?` inside string literals or
  backtick-quoted identifiers is left alone.

## Examples

```go
db.Query(`SELECT * FROM t WHERE x = ? AND y = ?`, 1, 2)
db.Query(`SELECT * FROM t WHERE x = @a AND y = @b`,
    sql.Named("a", 1), sql.Named("b", 2))
db.Query(`SELECT * FROM t WHERE x = @a AND y = ?`,
    sql.Named("a", 1), 2)
```

## Edge cases

- An args-less `PrepareContext` for a query that contains both `?`
  and `@name` still receives the rewrite at Prepare time, so the
  prepared SQL is purely named. Bind-time arg coverage is deferred
  to Exec / Query.
- An arg with a Name that does not appear in the query is silently
  ignored by the binder. An arg missing for a referenced name raises
  the underlying SQLite "no such column / parameter" error at
  Exec / Query.

## Reference (upstream)

- BigQuery parameterised queries: https://cloud.google.com/bigquery/docs/parameterized-queries
- database/sql NamedArg: https://pkg.go.dev/database/sql#NamedArg

## References

- BigQuery parameterised queries: https://cloud.google.com/bigquery/docs/parameterized-queries
- database/sql NamedArg: https://pkg.go.dev/database/sql#NamedArg
