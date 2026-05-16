---
name: ST_INTERSECTION
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Sutherland–Hodgman planar clip for polygon intersections; point/line intersections via membership filter. Runtime entry: BindStIntersection in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_intersection
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_intersection.yaml
---

# ST_INTERSECTION

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

## `ST_INTERSECTION`

```googlesql
ST_INTERSECTION(geography_1, geography_2)
```

**Description**

Returns a `GEOGRAPHY` that represents the point set
intersection of the two input `GEOGRAPHY`s. Thus,
every point in the intersection appears in both `geography_1` and `geography_2`.

If the two input `GEOGRAPHY`s are disjoint, that is,
there are no points that appear in both input `geometry_1` and `geometry_2`,
then an empty `GEOGRAPHY` is returned.

See [ST_INTERSECTS][st-intersects], [ST_DISJOINT][st-disjoint] for related
predicate functions.

**Return type**

`GEOGRAPHY`

[st-intersects]: #st_intersects

[st-disjoint]: #st_disjoint

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
