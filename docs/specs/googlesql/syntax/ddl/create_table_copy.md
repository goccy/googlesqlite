---
name: CREATE_TABLE_COPY
dialect: googlesql
category: syntax/ddl
status: partial
notes: |
  `FOR SYSTEM_TIME AS OF` is rejected. googlesqlite stores only the
  current state of a table, so there is no earlier version to read;
  accepting the clause would silently return present-day rows for a
  point-in-time request. Plain queries reject the clause too, through
  its own LanguageFeature gate.
source_url: docs/third_party/googlesql-docs/resolved_ast.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/resolved_ast.md
last_synced: 2026-08-19
testdata: testdata/specs/googlesql/syntax/ddl/create_table_copy.yaml
---

# `CREATE TABLE ... COPY`

## Summary

Creates a new table as a full copy of an existing table — both its
schema and its rows.

## Signatures

- `CREATE [OR REPLACE] [TEMP] TABLE [IF NOT EXISTS] <name> COPY <source> [OPTIONS (...)]`

## Behavior

- The whole schema is inherited from `<source>`: column names, types,
  `NOT NULL`, `DEFAULT` and the primary key.
- All rows of `<source>` are copied into the new table.
- `OPTIONS` given explicitly win; any option not named falls through
  from the source table.
- `COPY` and `CLONE` differ upstream only in expected cost — `CLONE` is
  documented as cheap and typically O(1), `COPY` as a full copy.
  googlesqlite materialises both the same way.

## Examples

```sql
CREATE TABLE dst COPY src;                            -- schema + rows
CREATE TABLE dst COPY src OPTIONS(description='d');   -- override one option
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

- `IF NOT EXISTS` makes the whole statement a no-op when the target
  exists. The row copy is a separate `INSERT` that SQLite's own
  `IF NOT EXISTS` does not guard, so re-running must not append a
  second set of rows.
- `FOR SYSTEM_TIME AS OF` is rejected; see the frontmatter notes.
- Copying a table into itself is rejected. `CREATE OR REPLACE` drops the
  target first, so an unguarded copy would read from the table it has
  just emptied and silently leave it with no rows. The check folds case,
  because SQLite resolves table names case-insensitively.

## Reference (upstream)

The grammar is recorded upstream under `ResolvedCreateTableStmt`:

```text
CREATE [TEMP] TABLE <name>
[(column schema, ...) | LIKE <name_path> |
    {CLONE|COPY} <name_path>
        [FOR SYSTEM_TIME AS OF <expr>]
        [WHERE <expr>]]
```

Upstream on option inheritance: "If the OPTIONS clause is explicitly
specified, the option values are intended to be used for the created or
replaced table. If any OPTION is unspecified, the corresponding option
from the source table will be used instead."

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/resolved_ast.md`.
