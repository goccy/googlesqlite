---
name: ST_MAKEPOLYGONORIENTED
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Preserves vertex orientation; identical to ST_MAKEPOLYGON in this driver. Runtime entry: BindStMakePolygonOriented in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_makepolygonoriented
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_makepolygonoriented.yaml
---

# ST_MAKEPOLYGONORIENTED

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

## `ST_MAKEPOLYGONORIENTED`

```googlesql
ST_MAKEPOLYGONORIENTED(array_of_geography)
```

**Description**

Like `ST_MAKEPOLYGON`, but the vertex ordering of each input linestring
determines the orientation of each polygon ring. The orientation of a polygon
ring defines the interior of the polygon as follows: if someone walks along the
boundary of the polygon in the order of the input vertices, the interior of the
polygon is on the left. This applies for each polygon ring provided.

This variant of the polygon constructor is more flexible since
`ST_MAKEPOLYGONORIENTED` can construct a polygon such that the interior is on
either side of the polygon ring. However, proper orientation of polygon rings is
critical in order to construct the desired polygon.

If the input `ARRAY` or any element in the `ARRAY` is `NULL`,
`ST_MAKEPOLYGONORIENTED` returns `NULL`.

NOTE: The input argument for `ST_MAKEPOLYGONORIENTED` may contain an empty
`GEOGRAPHY`. `ST_MAKEPOLYGONORIENTED` interprets an empty `GEOGRAPHY` as having
an empty linestring, which will create a full loop: that is, a polygon that
covers the entire Earth.

**Constraints**

Together, the input rings must form a valid polygon:

+   The polygon shell must cover each of the polygon holes.
+   There must be only one polygon shell, which must to be the first input ring.
    This implies that polygon holes can't be nested.
+   Polygon rings may only intersect in a vertex on the boundary of both rings.

Every edge must span strictly less than 180 degrees.

`ST_MAKEPOLYGONORIENTED` relies on the ordering of the input vertices of each
linestring to determine the orientation of the polygon. This applies to the
polygon shell and any polygon holes. `ST_MAKEPOLYGONORIENTED` expects all
polygon holes to have the opposite orientation of the shell. See
[`ST_MAKEPOLYGON`][st-makepolygon] for an alternate polygon constructor, and
other constraints on building a valid polygon.

NOTE: Due to the GoogleSQL snapping process, edges with a sufficiently
short length will be discarded and the two endpoints will be snapped to a single
point. Therefore, it's possible that vertices in a linestring may be snapped
together such that one or more edge disappears. Hence, it's possible that a
polygon hole that's sufficiently small may disappear, or the resulting
`GEOGRAPHY` may contain only a line or a point.

**Return type**

`GEOGRAPHY`

[st-makepolygon]: #st_makepolygon

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
