---
name: ST_EXTENT
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  ST_EXTENT is registered as an aggregate via
  internal/functions/geography/st_extent_agg.go. Step accumulates
  every vertex contributed by each input geography (POINT EMPTY
  contributes no vertices). Done produces a
  STRUCT<xmin FLOAT64, ymin FLOAT64, xmax FLOAT64, ymax FLOAT64>;
  if any contributing shape crosses the antimeridian, every
  negative longitude in the aggregation is reinterpreted as
  `lng + 360` before the min/max scan so the resulting box is the
  narrow span. That matches upstream's Example 1, where a polygon
  in (-125..-117) co-aggregated with an antimeridian-crossing
  polygon at (172..-141) yields xmin=172, xmax=243 (= -117 + 360).
  Empty aggregation returns NULL.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_extent
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_extent.yaml
---

# ST_EXTENT

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

## `ST_EXTENT`

```googlesql
ST_EXTENT(geography_expression)
```

**Description**

Returns a `STRUCT` that represents the bounding box for the set of input
`GEOGRAPHY` values. The bounding box is the minimal rectangle that encloses the
geography. The edges of the rectangle follow constant lines of longitude and
latitude.

Caveats:

+ Returns `NULL` if all the inputs are `NULL` or empty geographies.
+ The bounding box might cross the antimeridian if this allows for a smaller
  rectangle. In this case, the bounding box has one of its longitudinal bounds
  outside of the [-180, 180] range, so that `xmin` is smaller than the eastmost
  value `xmax`.
+ If the longitude span of the bounding box is larger than or equal to 180
  degrees, the function returns the bounding box with the longitude range of
  [-180, 180].

**Return type**

`STRUCT<xmin DOUBLE, ymin DOUBLE, xmax DOUBLE, ymax DOUBLE>`.

Bounding box parts:

+ `xmin`: The westmost constant longitude line that bounds the rectangle.
+ `xmax`: The eastmost constant longitude line that bounds the rectangle.
+ `ymin`: The minimum constant latitude line that bounds the rectangle.
+ `ymax`: The maximum constant latitude line that bounds the rectangle.

**Example**

```googlesql
WITH data AS (
  SELECT 1 id, ST_GEOGFROMTEXT('POLYGON((-125 48, -124 46, -117 46, -117 49, -125 48))') g
  UNION ALL
  SELECT 2 id, ST_GEOGFROMTEXT('POLYGON((172 53, -130 55, -141 70, 172 53))') g
  UNION ALL
  SELECT 3 id, ST_GEOGFROMTEXT('POINT EMPTY') g
)
SELECT ST_EXTENT(g) AS box
FROM data

/*----------------------------------------------+
 | box                                          |
 +----------------------------------------------+
 | {xmin:172, ymin:46, xmax:243, ymax:70}       |
 +----------------------------------------------*/
```

[`ST_BOUNDINGBOX`][st-boundingbox] for the non-aggregate version of `ST_EXTENT`.

[st-boundingbox]: #st_boundingbox

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
