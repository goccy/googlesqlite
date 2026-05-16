---
name: ST_BOUNDARY
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Boundary of a LINESTRING (start+end POINTs) / POLYGON (rings as MULTILINESTRING). Runtime entry: BindStBoundary in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_boundary
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_boundary.yaml
---

# ST_BOUNDARY

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

## `ST_BOUNDARY`

```googlesql
ST_BOUNDARY(geography_expression)
```

**Description**

Returns a single `GEOGRAPHY` that contains the union
of the boundaries of each component in the given input
`GEOGRAPHY`.

The boundary of each component of a `GEOGRAPHY` is
defined as follows:

+   The boundary of a point is empty.
+   The boundary of a linestring consists of the endpoints of the linestring.
+   The boundary of a polygon consists of the linestrings that form the polygon
    shell and each of the polygon's holes.

**Return type**

`GEOGRAPHY`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
