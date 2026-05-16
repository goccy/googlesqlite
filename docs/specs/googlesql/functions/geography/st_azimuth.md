---
name: ST_AZIMUTH
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Geodesic bearing from point A to point B in radians clockwise from north. Runtime entry: BindStAzimuth in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_azimuth
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_azimuth.yaml
---

# ST_AZIMUTH

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

## `ST_AZIMUTH`

```googlesql
ST_AZIMUTH(point_geography_1, point_geography_2)
```

**Description**

Takes two point `GEOGRAPHY` values, and returns the azimuth of the line segment
formed by points 1 and 2. The azimuth is the angle in radians measured between
the line from point 1 facing true North to the line segment from point 1 to
point 2.

The positive angle is measured clockwise on the surface of a sphere. For
example, the azimuth for a line segment:

+   Pointing North is `0`
+   Pointing East is `PI/2`
+   Pointing South is `PI`
+   Pointing West is `3PI/2`

`ST_AZIMUTH` has the following edge cases:

+   If the two input points are the same, returns `NULL`.
+   If the two input points are exactly antipodal, returns `NULL`.
+   If either of the input geographies aren't single points or are the empty
    geography, throws an error.

**Return type**

`DOUBLE`

**Example**

```googlesql
WITH geos AS (
  SELECT 1 id, ST_GEOGPOINT(1, 0) AS geo1, ST_GEOGPOINT(0, 0) AS geo2 UNION ALL
  SELECT 2, ST_GEOGPOINT(0, 0), ST_GEOGPOINT(1, 0) UNION ALL
  SELECT 3, ST_GEOGPOINT(0, 0), ST_GEOGPOINT(0, 1) UNION ALL
  -- identical
  SELECT 4, ST_GEOGPOINT(0, 0), ST_GEOGPOINT(0, 0) UNION ALL
  -- antipode
  SELECT 5, ST_GEOGPOINT(-30, 0), ST_GEOGPOINT(150, 0) UNION ALL
  -- nulls
  SELECT 6, ST_GEOGPOINT(0, 0), NULL UNION ALL
  SELECT 7, NULL, ST_GEOGPOINT(0, 0))
SELECT ST_AZIMUTH(geo1, geo2) AS azimuth FROM geos ORDER BY id;

/*--------------------+
 | azimuth            |
 +--------------------+
 | 4.71238898038469   |
 | 1.5707963267948966 |
 | 0                  |
 | NULL               |
 | NULL               |
 | NULL               |
 | NULL               |
 +--------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
