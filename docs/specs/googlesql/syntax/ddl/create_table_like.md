---
name: CREATE_TABLE_LIKE
dialect: googlesql
category: syntax/ddl
status: implemented
source_url: docs/third_party/googlesql-docs/resolved_ast.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/resolved_ast.md
last_synced: 2026-08-19
testdata: testdata/specs/googlesql/syntax/ddl/create_table_like.yaml
---

# `CREATE TABLE ... LIKE`

## Summary

Creates a new table with the same schema as an existing table, without
copying any of its rows.

## Signatures

- `CREATE [OR REPLACE] [TEMP] TABLE [IF NOT EXISTS] <name> LIKE <source>`

## Behavior

- The new table's columns — names, types, `NOT NULL`, `DEFAULT` and the
  primary key — are inherited from `<source>`.
- No rows are copied. The new table is empty.
- `<source>` must exist; an unknown name is an analysis error.

## Examples

```sql
CREATE TABLE dst LIKE src;   -- dst has src's schema and no rows
```

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

- The analyzer materialises `LIKE` into a column list carrying names and
  types only. googlesqlite recovers `NOT NULL`, `DEFAULT` and the
  primary key from the source table's stored schema so the inherited
  constraints survive.
- Unlike `COPY` / `CLONE`, `LIKE` does not inherit the source table's
  `OPTIONS`. Upstream states option inheritance only for the
  `CLONE` / `COPY` forms.

## Reference (upstream)

The grammar is recorded upstream under `ResolvedCreateTableStmt`:

```text
CREATE [TEMP] TABLE <name>
[(column schema, ...) | LIKE <name_path> |
    {CLONE|COPY} <name_path>
        [FOR SYSTEM_TIME AS OF <expr>]
        [WHERE <expr>]]
```

## References

Apache 2.0 derivative of `docs/third_party/googlesql-docs/resolved_ast.md`.
