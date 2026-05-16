---
name: ST_LINEINTERPOLATEPOINT
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Point at the given fraction along a LINESTRING's geodesic length using s2.Polyline.Interpolate. Runtime entry: BindStLineInterpolatePoint in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_lineinterpolatepoint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_lineinterpolatepoint.yaml
---

# ST_LINEINTERPOLATEPOINT

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

## `ST_LINEINTERPOLATEPOINT`

```googlesql
ST_LINEINTERPOLATEPOINT(linestring_geography, fraction)
```

**Description**

Gets a point at a specific fraction in a linestring `GEOGRAPHY` value.

**Definitions**

+  `linestring_geography`: A linestring `GEOGRAPHY` on which the target point
    is located.
+  `fraction`: A `DOUBLE` value that represents a fraction
    along the linestring `GEOGRAPHY` where the target point is located.
    This should be an inclusive value between `0` (start of the
    linestring) and `1` (end of the linestring).

**Details**

+   Returns `NULL` if any input argument is `NULL`.
+   Returns an empty geography if `linestring_geography` is an empty geography.
+   Returns an error if `linestring_geography` isn't a linestring or an empty
    geography, or if `fraction` is outside the `[0, 1]` range.

**Return Type**

`GEOGRAPHY`

**Example**

The following query returns a few points on a linestring. Notice that the
 midpoint of the linestring `LINESTRING(1 1, 5 5)` is slightly different from
 `POINT(3 3)` because the `GEOGRAPHY` type uses geodesic line segments.

```googlesql
WITH fractions AS (
    SELECT 0 AS fraction UNION ALL
    SELECT 0.5 UNION ALL
    SELECT 1 UNION ALL
    SELECT NULL
  )
SELECT
  fraction,
  ST_LINEINTERPOLATEPOINT(ST_GEOGFROMTEXT('LINESTRING(1 1, 5 5)'), fraction)
    AS point
FROM fractions

/*-------------+-------------------------------------------+
 | fraction    | point                                     |
 +-------------+-------------------------------------------+
 | 0           | POINT(1 1)                                |
 | 0.5         | POINT(2.99633827268976 3.00182528336078)  |
 | 1           | POINT(5 5)                                |
 | NULL        | NULL                                      |
 +-------------+-------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
