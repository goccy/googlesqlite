---
name: CREATE_TABLE_CLONE
dialect: googlesql
category: syntax/ddl
status: partial
notes: |
  `FOR SYSTEM_TIME AS OF` is rejected. googlesqlite stores only the
  current state of a table, so there is no earlier version to read;
  accepting the clause would silently return present-day rows for a
  point-in-time request. Plain queries reject the clause too, through
  its own LanguageFeature gate.

  `CLONE` is materialised as a full copy. Upstream describes it as a
  cheap, typically O(1) operation, which googlesqlite has no storage
  layer to implement; the observable result is identical to `COPY`.
source_url: docs/third_party/googlesql-docs/resolved_ast.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/resolved_ast.md
last_synced: 2026-08-19
testdata: testdata/specs/googlesql/syntax/ddl/create_table_clone.yaml
---

# `CREATE TABLE ... CLONE`

## Summary

Creates a new table from an existing one, inheriting its schema and its
rows, optionally restricted by a `WHERE` clause.

## Signatures

- `CREATE [OR REPLACE] [TEMP] TABLE [IF NOT EXISTS] <name> CLONE <source> [WHERE <expr>] [OPTIONS (...)]`

## Behavior

- The whole schema is inherited from `<source>`: column names, types,
  `NOT NULL`, `DEFAULT` and the primary key.
- Rows of `<source>` are copied. With a `WHERE` clause only the matching
  rows are.
- `OPTIONS` given explicitly win; any option not named falls through
  from the source table.

## Examples

```sql
CREATE TABLE dst CLONE src;                -- schema + all rows
CREATE TABLE dst CLONE src WHERE a > 1;    -- schema + matching rows
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

- A `WHERE` clause makes the analyzer wrap the source table scan in a
  filter scan. Upstream permits exactly that one wrapper and no other
  scan type, so the source table stays directly identifiable.
- `IF NOT EXISTS` makes the whole statement a no-op when the target
  exists, including the row copy.
- `FOR SYSTEM_TIME AS OF` is rejected; see the frontmatter notes.
- Cloning a table into itself is rejected, for the same reason
  `COPY` is: `CREATE OR REPLACE` empties the target before the copy
  reads from it.
- A view may be used as the source; its rows are materialised into the
  new table.
- The `CREATE` and the row copy are separate statements. If the copy
  fails — a `WHERE` clause that errors at evaluation time, say — the
  created table is dropped again so the name stays reusable.

## Reference (upstream)

The grammar is recorded upstream under `ResolvedCreateTableStmt`:

```text
CREATE [TEMP] TABLE <name>
[(column schema, ...) | LIKE <name_path> |
    {CLONE|COPY} <name_path>
        [FOR SYSTEM_TIME AS OF <expr>]
        [WHERE <expr>]]
```

Upstream on the source scan: "ResolvedTableScan will represent the
source table, with an optional for_system_time_expr. The
ResolvedTableScan may be wrapped inside a ResolvedFilterScan if the
source table has a where clause. No other Scan types are allowed here."

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/resolved_ast.md`.
