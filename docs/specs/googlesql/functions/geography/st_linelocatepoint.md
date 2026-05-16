---
name: ST_LINELOCATEPOINT
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_linelocatepoint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_linelocatepoint.yaml
---

# ST_LINELOCATEPOINT

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

## `ST_LINELOCATEPOINT`

```googlesql
ST_LINELOCATEPOINT(linestring_geography, point_geography)
```

**Description**

Gets a section of a linestring between the start point and a selected point (a
point on the linestring closest to the `point_geography` argument). Returns the
percentage that this section represents in the linestring.

Details:

+   To select a point on the linestring `GEOGRAPHY` (`linestring_geography`),
    this function takes a point `GEOGRAPHY` (`point_geography`) and finds the
    [closest point][st-closestpoint] to it on the linestring.
+   If two points on `linestring_geography` are an equal distance away from
    `point_geography`, it isn't guaranteed which one will be selected.
+   The return value is an inclusive value between 0 and 1 (0-100%).
+   If the selected point is the start point on the linestring, function returns
    0 (0%).
+   If the selected point is the end point on the linestring, function returns 1
    (100%).

`NULL` and error handling:

+   Returns `NULL` if any input argument is `NULL`.
+   Returns an error if `linestring_geography` isn't a linestring or if
    `point_geography` isn't a point. Use the `SAFE` prefix
    to obtain `NULL` for invalid input instead of an error.

**Return Type**

`DOUBLE`

**Examples**

```googlesql
WITH geos AS (
    SELECT ST_GEOGPOINT(0, 0) AS point UNION ALL
    SELECT ST_GEOGPOINT(1, 0) UNION ALL
    SELECT ST_GEOGPOINT(1, 1) UNION ALL
    SELECT ST_GEOGPOINT(2, 2) UNION ALL
    SELECT ST_GEOGPOINT(3, 3) UNION ALL
    SELECT ST_GEOGPOINT(4, 4) UNION ALL
    SELECT ST_GEOGPOINT(5, 5) UNION ALL
    SELECT ST_GEOGPOINT(6, 5) UNION ALL
    SELECT NULL
  )
SELECT
  point AS input_point,
  ST_LINELOCATEPOINT(ST_GEOGFROMTEXT('LINESTRING(1 1, 5 5)'), point)
    AS percentage_from_beginning
FROM geos

/*-------------+---------------------------+
 | input_point | percentage_from_beginning |
 +-------------+---------------------------+
 | POINT(0 0)  | 0                         |
 | POINT(1 0)  | 0                         |
 | POINT(1 1)  | 0                         |
 | POINT(2 2)  | 0.25015214685147907       |
 | POINT(3 3)  | 0.5002284283637185        |
 | POINT(4 4)  | 0.7501905913884388        |
 | POINT(5 5)  | 1                         |
 | POINT(6 5)  | 1                         |
 | NULL        | NULL                      |
 +-------------+---------------------------*/
```

[st-closestpoint]: #st_closestpoint

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
