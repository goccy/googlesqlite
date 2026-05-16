---
name: INFORMATION_SCHEMA_TABLE_OPTIONS
dialect: googlesql
category: information_schema
status: implemented
source_url: https://cloud.google.com/bigquery/docs/information-schema-table-options
upstream_url: https://cloud.google.com/bigquery/docs/information-schema-table-options
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/information_schema/table_options.yaml
---

# `INFORMATION_SCHEMA.TABLE_OPTIONS`

## Summary

`<dataset>.INFORMATION_SCHEMA.TABLE_OPTIONS` exposes one row per
declared option on a table — `description`, `labels`, etc.

## Columns (selected)

- `table_name STRING`
- `option_name STRING`
- `option_value STRING`

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

Documentation source: <https://cloud.google.com/bigquery/docs/information-schema-table-options>.
