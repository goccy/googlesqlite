---
name: INFORMATION_SCHEMA_SCHEMATA
dialect: googlesql
category: information_schema
status: implemented
source_url: https://cloud.google.com/bigquery/docs/information-schema-schemata
upstream_url: https://cloud.google.com/bigquery/docs/information-schema-schemata
last_synced: 2026-05-12
testdata: testdata/specs/googlesql/information_schema/schemata.yaml
---

# `INFORMATION_SCHEMA.SCHEMATA`

## Summary

`<project>.INFORMATION_SCHEMA.SCHEMATA` exposes one row per dataset
visible to the named project.

## Columns (selected)

- `catalog_name STRING`
- `schema_name STRING`
- `schema_owner STRING`

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

Documentation source: <https://cloud.google.com/bigquery/docs/information-schema-schemata>.
