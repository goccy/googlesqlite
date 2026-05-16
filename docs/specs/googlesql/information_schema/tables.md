---
name: INFORMATION_SCHEMA_TABLES
dialect: googlesql
category: information_schema
status: implemented
source_url: https://cloud.google.com/bigquery/docs/information-schema-tables
upstream_url: https://cloud.google.com/bigquery/docs/information-schema-tables
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/information_schema/tables.yaml
---

# `INFORMATION_SCHEMA.TABLES`

## Summary

`<dataset>.INFORMATION_SCHEMA.TABLES` exposes one row per table in the
named dataset. Each row records the table's schema name, table name,
and table type (`BASE TABLE`, `VIEW`, etc.).

## Columns (selected)

- `table_name STRING`
- `table_schema STRING`
- `table_type STRING`

## Signatures

See the upstream reference linked at the bottom of this spec.

## Behavior

See the upstream reference linked at the bottom of this spec.

## Examples

See the upstream reference linked at the bottom of this spec and the testdata YAML.

## Edge cases

Covered by the testdata YAML linked in the frontmatter.

## Reference (upstream)

See the `upstream_url` and `source_url` fields in this spec's frontmatter.

## References

Documentation source: <https://cloud.google.com/bigquery/docs/information-schema-tables>.
