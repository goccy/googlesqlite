---
name: ST_PERIMETER
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Geodesic perimeter of a (multi)polygon in meters via per-loop edge-distance accumulation. Runtime entry: BindStPerimeter in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_perimeter
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_perimeter.yaml
---

# ST_PERIMETER

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

## `ST_PERIMETER`

```googlesql
ST_PERIMETER(geography_expression[, use_spheroid])
```

**Description**

Returns the length in meters of the boundary of the polygons in the input
`GEOGRAPHY`.

If `geography_expression` is a point or a line, returns zero. If
`geography_expression` is a collection, returns the perimeter of the polygons
in the collection; if the collection doesn't contain polygons, returns zero.

The optional `use_spheroid` parameter determines how this function measures
distance. If `use_spheroid` is `FALSE`, the function measures distance on the
surface of a perfect sphere.

The `use_spheroid` parameter currently only supports
the value `FALSE`. The default value of `use_spheroid` is `FALSE`.

**Return type**

`DOUBLE`

[wgs84-link]: https://en.wikipedia.org/wiki/World_Geodetic_System

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
