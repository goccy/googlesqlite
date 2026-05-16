---
name: ST_COVERS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Same vertex-membership predicate as ST_CONTAINS for our supported shapes (boundary points are reported as contained by s2). Runtime entry: BindStCovers in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_covers
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_covers.yaml
---

# ST_COVERS

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

## `ST_COVERS`

```googlesql
ST_COVERS(geography_1, geography_2)
```

**Description**

Returns `FALSE` if `geography_1` or `geography_2` is empty.
Returns `TRUE` if no points of `geography_2` lie in the exterior of
`geography_1`.

**Return type**

`BOOL`

**Example**

The following query tests whether the polygon `POLYGON((1 1, 20 1, 10 20, 1 1))`
covers each of the three points `(0, 0)`, `(1, 1)`, and `(10, 10)`, which lie
on the exterior, the boundary, and the interior of the polygon respectively.

```googlesql
SELECT
  ST_GEOGPOINT(i, i) AS p,
  ST_COVERS(ST_GEOGFROMTEXT('POLYGON((1 1, 20 1, 10 20, 1 1))'),
            ST_GEOGPOINT(i, i)) AS `covers`
FROM UNNEST([0, 1, 10]) AS i;

/*--------------+--------+
 | p            | covers |
 +--------------+--------+
 | POINT(0 0)   | FALSE  |
 | POINT(1 1)   | TRUE   |
 | POINT(10 10) | TRUE   |
 +--------------+--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
