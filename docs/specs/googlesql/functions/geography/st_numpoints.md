---
name: ST_NUMPOINTS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Total point count across the geography. Runtime entry: BindStNumPoints in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_numpoints
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_numpoints.yaml
---

# ST_NUMPOINTS

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

## `ST_NUMPOINTS`

```googlesql
ST_NUMPOINTS(geography_expression)
```

**Description**

Returns the number of vertices in the input
`GEOGRAPHY`. This includes the number of points, the
number of linestring vertices, and the number of polygon vertices.

NOTE: The first and last vertex of a polygon ring are counted as distinct
vertices.

**Return type**

`INT64`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
