---
name: ST_ANGLE
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_angle
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_angle.yaml
---

# ST_ANGLE

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

## `ST_ANGLE`

```googlesql
ST_ANGLE(point_geography_1, point_geography_2, point_geography_3)
```

**Description**

Takes three point `GEOGRAPHY` values, which represent two intersecting lines.
Returns the angle between these lines. Point 2 and point 1 represent the first
line and point 2 and point 3 represent the second line. The angle between
these lines is in radians, in the range `[0, 2pi)`. The angle is measured
clockwise from the first line to the second line.

`ST_ANGLE` has the following edge cases:

+ If points 2 and 3 are the same, returns `NULL`.
+ If points 2 and 1 are the same, returns `NULL`.
+ If points 2 and 3 are exactly antipodal, returns `NULL`.
+ If points 2 and 1 are exactly antipodal, returns `NULL`.
+ If any of the input geographies aren't single points or are the empty
  geography, then throws an error.

**Return type**

`DOUBLE`

**Example**

```googlesql
WITH geos AS (
  SELECT 1 id, ST_GEOGPOINT(1, 0) geo1, ST_GEOGPOINT(0, 0) geo2, ST_GEOGPOINT(0, 1) geo3 UNION ALL
  SELECT 2 id, ST_GEOGPOINT(0, 0), ST_GEOGPOINT(1, 0), ST_GEOGPOINT(0, 1) UNION ALL
  SELECT 3 id, ST_GEOGPOINT(1, 0), ST_GEOGPOINT(0, 0), ST_GEOGPOINT(1, 0) UNION ALL
  SELECT 4 id, ST_GEOGPOINT(1, 0) geo1, ST_GEOGPOINT(0, 0) geo2, ST_GEOGPOINT(0, 0) geo3 UNION ALL
  SELECT 5 id, ST_GEOGPOINT(0, 0), ST_GEOGPOINT(-30, 0), ST_GEOGPOINT(150, 0) UNION ALL
  SELECT 6 id, ST_GEOGPOINT(0, 0), NULL, NULL UNION ALL
  SELECT 7 id, NULL, ST_GEOGPOINT(0, 0), NULL UNION ALL
  SELECT 8 id, NULL, NULL, ST_GEOGPOINT(0, 0))
SELECT ST_ANGLE(geo1,geo2,geo3) AS angle FROM geos ORDER BY id;

/*---------------------+
 | angle               |
 +---------------------+
 | 4.71238898038469    |
 | 0.78547432161873854 |
 | 0                   |
 | NULL                |
 | NULL                |
 | NULL                |
 | NULL                |
 | NULL                |
 +---------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
