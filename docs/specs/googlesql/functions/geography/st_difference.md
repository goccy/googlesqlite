---
name: ST_DIFFERENCE
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  ST_DIFFERENCE has a fast-path when B is entirely contained in A
  with no boundary overlap (every vertex of B's outer ring is inside
  A and no vertex of A is inside B): the result is A with B's outer
  ring appended as a reversed inner ring (per OGC SFS the inner
  ring winds opposite to the outer). The general convex-hull
  fallback is kept for partial-overlap cases until proper polygon-
  difference clipping lands.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_difference
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_difference.yaml
---

# ST_DIFFERENCE

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

## `ST_DIFFERENCE`

```googlesql
ST_DIFFERENCE(geography_1, geography_2)
```

**Description**

Returns a `GEOGRAPHY` that represents the point set
difference of `geography_1` and `geography_2`. Therefore, the result consists of
the part of `geography_1` that doesn't intersect with `geography_2`.

If `geometry_1` is completely contained in `geometry_2`, then `ST_DIFFERENCE`
returns an empty `GEOGRAPHY`.

**Constraints**

The underlying geometric objects that a GoogleSQL
`GEOGRAPHY` represents correspond to a *closed* point
set. Therefore, `ST_DIFFERENCE` is the closure of the point set difference of
`geography_1` and `geography_2`. This implies that if `geography_1` and
`geography_2` intersect, then a portion of the boundary of `geography_2` could
be in the difference.

**Return type**

`GEOGRAPHY`

**Example**

The following query illustrates the difference between `geog1`, a larger polygon
`POLYGON((0 0, 10 0, 10 10, 0 0))` and `geog2`, a smaller polygon
`POLYGON((4 2, 6 2, 8 6, 4 2))` that intersects with `geog1`. The result is
`geog1` with a hole where `geog2` intersects with it.

```googlesql
SELECT
  ST_DIFFERENCE(
      ST_GEOGFROMTEXT('POLYGON((0 0, 10 0, 10 10, 0 0))'),
      ST_GEOGFROMTEXT('POLYGON((4 2, 6 2, 8 6, 4 2))')
  );

/*--------------------------------------------------------+
 | difference_of_geog1_and_geog2                          |
 +--------------------------------------------------------+
 | POLYGON((0 0, 10 0, 10 10, 0 0), (8 6, 6 2, 4 2, 8 6)) |
 +--------------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
