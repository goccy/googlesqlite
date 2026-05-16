---
name: ST_BOUNDINGBOX
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStBoundingBox returns the axis-aligned bounding box as a
  STRUCT<xmin FLOAT64, ymin FLOAT64, xmax FLOAT64, ymax FLOAT64>.
  POINT EMPTY returns NULL — the WKT parser now models the
  empty-point sentinel distinctly from POINT(0 0). When the geometry
  crosses the antimeridian (any edge with longitude delta > 180°),
  negative longitudes are unwrapped by +360° before the min/max
  scan, so a polygon hugging the pole near 180° / -180° produces a
  narrow box (e.g. the Example-2 polygon yields xmin=172, xmax=230).
  
  `oriented => TRUE` is honoured: when the polygon's outer ring is
  clockwise (signed Shoelace area on antimeridian-unwrapped
  longitudes is negative), the polygon is marked `inverted` — the
  interior covers the entire globe minus the ring's small enclosed
  region — and ST_BOUNDINGBOX returns the full-globe box
  `{xmin:-180, ymin:-90, xmax:180, ymax:90}`. The `inverted` flag is
  also encoded across the value codec round-trip via an `INVERTED `
  sentinel prefix on the WKT body.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_boundingbox
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_boundingbox.yaml
---

# ST_BOUNDINGBOX

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

## `ST_BOUNDINGBOX`

```googlesql
ST_BOUNDINGBOX(geography_expression)
```

**Description**

Returns a `STRUCT` that represents the bounding box for the specified geography.
The bounding box is the minimal rectangle that encloses the geography. The edges
of the rectangle follow constant lines of longitude and latitude.

Caveats:

+ Returns `NULL` if the input is `NULL` or an empty geography.
+ The bounding box might cross the antimeridian if this allows for a smaller
  rectangle. In this case, the bounding box has one of its longitudinal bounds
  outside of the [-180, 180] range, so that `xmin` is smaller than the eastmost
  value `xmax`.

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
  UNION ALL
  SELECT 4 id, ST_GEOGFROMTEXT('POLYGON((172 53, -141 70, -130 55, 172 53))', oriented => TRUE)
)
SELECT id, ST_BOUNDINGBOX(g) AS box
FROM data

/*----+------------------------------------------+
 | id | box                                      |
 +----+------------------------------------------+
 | 1  | {xmin:-125, ymin:46, xmax:-117, ymax:49} |
 | 2  | {xmin:172, ymin:53, xmax:230, ymax:70}   |
 | 3  | NULL                                     |
 | 4  | {xmin:-180, ymin:-90, xmax:180, ymax:90} |
 +----+------------------------------------------*/
```

See [`ST_EXTENT`][st-extent] for the aggregate version of `ST_BOUNDINGBOX`.

[st-extent]: #st_extent

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
