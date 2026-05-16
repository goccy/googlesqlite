---
name: ST_CONTAINS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  True when every vertex of B lies inside A. Implemented via per-vertex membership against s2.Polygon.ContainsPoint. Runtime entry: BindStContains in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_contains
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_contains.yaml
---

# ST_CONTAINS

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

## `ST_CONTAINS`

```googlesql
ST_CONTAINS(geography_1, geography_2)
```

**Description**

Returns `TRUE` if no point of `geography_2` is outside `geography_1`, and
the interiors intersect; returns `FALSE` otherwise.

NOTE: A `GEOGRAPHY` *does not* contain its own
boundary. Compare with [`ST_COVERS`][st_covers].

**Return type**

`BOOL`

**Example**

The following query tests whether the polygon `POLYGON((1 1, 20 1, 10 20, 1 1))`
contains each of the three points `(0, 0)`, `(1, 1)`, and `(10, 10)`, which lie
on the exterior, the boundary, and the interior of the polygon respectively.

```googlesql
SELECT
  ST_GEOGPOINT(i, i) AS p,
  ST_CONTAINS(ST_GEOGFROMTEXT('POLYGON((1 1, 20 1, 10 20, 1 1))'),
              ST_GEOGPOINT(i, i)) AS `contains`
FROM UNNEST([0, 1, 10]) AS i;

/*--------------+----------+
 | p            | contains |
 +--------------+----------+
 | POINT(0 0)   | FALSE    |
 | POINT(1 1)   | FALSE    |
 | POINT(10 10) | TRUE     |
 +--------------+----------*/
```

[st_covers]: #st_covers

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
