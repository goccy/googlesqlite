---
name: CREATE_SNAPSHOT_TABLE
dialect: googlesql
category: syntax/ddl
status: partial
notes: |
  googlesqlite materialises a snapshot as an ordinary table holding a
  copy of the source rows. It is neither point-in-time nor read-only:
  the driver keeps only the current state of a table, so there is no
  earlier version to pin, and it enforces no table-level access
  control. `FOR SYSTEM_TIME AS OF` is rejected for the same reason.

  Callers needing true snapshot isolation should not rely on this.
source_url: docs/third_party/googlesql-docs/resolved_ast.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/resolved_ast.md
last_synced: 2026-08-19
testdata: testdata/specs/googlesql/syntax/ddl/create_snapshot_table.yaml
---

# `CREATE SNAPSHOT TABLE`

## Summary

Creates a snapshot of an existing table, inheriting its schema and rows.

## Signatures

- `CREATE SNAPSHOT TABLE [IF NOT EXISTS] <name> CLONE <source> [OPTIONS (...)]`
- `DROP SNAPSHOT TABLE [IF EXISTS] <name>`

## Behavior

- The whole schema is inherited from `<source>`: column names, types,
  `NOT NULL`, `DEFAULT` and the primary key.
- All rows of `<source>` are copied into the snapshot.
- `OPTIONS` given explicitly win; any option not named falls through
  from the source table.
- `DROP SNAPSHOT TABLE` removes the snapshot. Since a snapshot is an
  ordinary table here, this is the same work as `DROP TABLE`.

## Examples

```sql
CREATE SNAPSHOT TABLE snap CLONE src;
DROP SNAPSHOT TABLE snap;
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

- `DROP SNAPSHOT TABLE IF EXISTS` on a missing table succeeds.
- The snapshot is writable and tracks nothing after creation; see the
  frontmatter notes for the divergence from a real snapshot table.

## Reference (upstream)

The grammar is recorded upstream under
`ResolvedCreateSnapshotTableStmt`:

```text
CREATE SNAPSHOT TABLE [IF NOT EXISTS] <name> [OPTIONS (...)]
CLONE <name_path>
        [FOR SYSTEM_TIME AS OF <expr>]
```

Upstream on inheritance: "By default, all fields (column names, types,
constraints, partition, clustering, options etc.) will be inherited from
the source table. If table options are explicitly set, the explicit
options will take precedence."

The `CREATE SNAPSHOT TABLE` section of
`docs/third_party/googlesql-docs/data-definition-language.md` reads
"Documentation is pending for this feature", so the resolved-AST record
above is the authoritative upstream description.

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/resolved_ast.md`.
