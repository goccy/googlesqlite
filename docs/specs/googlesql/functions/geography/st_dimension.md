---
name: ST_DIMENSION
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Topological dimension (0/1/2 for point/line/polygon, -1 for empty). Runtime entry: BindStDimension in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_dimension
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_dimension.yaml
---

# ST_DIMENSION

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

## `ST_DIMENSION`

```googlesql
ST_DIMENSION(geography_expression)
```

**Description**

Returns the dimension of the highest-dimensional element in the input
`GEOGRAPHY`.

The dimension of each possible element is as follows:

+   The dimension of a point is `0`.
+   The dimension of a linestring is `1`.
+   The dimension of a polygon is `2`.

If the input `GEOGRAPHY` is empty, `ST_DIMENSION`
returns `-1`.

**Return type**

`INT64`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
