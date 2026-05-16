---
name: INFORMATION_SCHEMA_COLUMNS
dialect: googlesql
category: information_schema
status: implemented
source_url: https://cloud.google.com/bigquery/docs/information-schema-columns
upstream_url: https://cloud.google.com/bigquery/docs/information-schema-columns
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/information_schema/columns.yaml
---

# `INFORMATION_SCHEMA.COLUMNS`

## Summary

`<dataset>.INFORMATION_SCHEMA.COLUMNS` exposes one row per column in
every table of the named dataset. Each row carries the column name,
its declared data type, the ordinal position within the parent table,
and whether the column is nullable.

## Columns (selected)

- `table_name STRING`
- `column_name STRING`
- `data_type STRING`
- `is_nullable STRING`
- `ordinal_position INT64`

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

Documentation source: <https://cloud.google.com/bigquery/docs/information-schema-columns>.
