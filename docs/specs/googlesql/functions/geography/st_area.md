---
name: ST_AREA
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Geodesic area of a (multi)polygon in square meters; uses s2.Polygon.Area() scaled by Earth radius squared. Runtime entry: BindStArea in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_area
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_area.yaml
---

# ST_AREA

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

## `ST_AREA`

```googlesql
ST_AREA(geography_expression[, use_spheroid])
```

**Description**

Returns the area in square meters covered by the polygons in the input
`GEOGRAPHY`.

If `geography_expression` is a point or a line, returns zero. If
`geography_expression` is a collection, returns the area of the polygons in the
collection; if the collection doesn't contain polygons, returns zero.

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
