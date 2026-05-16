---
name: ST_ISCOLLECTION
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Whether the geography is a MULTI* or GEOMETRYCOLLECTION. Runtime entry: BindStIsCollection in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_iscollection
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_iscollection.yaml
---

# ST_ISCOLLECTION

## Summary

(TBD — refine from the upstream reference below.)

## Signatures

(TBD)

## Behavior

(TBD)

## Examples

(TBD)

## Edge cases

(TBD)

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/geography_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ST_ISCOLLECTION`

```googlesql
ST_ISCOLLECTION(geography_expression)
```

**Description**

Returns `TRUE` if the total number of points, linestrings, and polygons is
greater than one.

An empty `GEOGRAPHY` isn't a collection.

**Return type**

`BOOL`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
