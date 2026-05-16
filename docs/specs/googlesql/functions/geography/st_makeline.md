---
name: ST_MAKELINE
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Joins ordered POINTs / LINESTRINGs into a single LINESTRING. Runtime entry: BindStMakeLine in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_makeline
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_makeline.yaml
---

# ST_MAKELINE

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

## `ST_MAKELINE`

```googlesql
ST_MAKELINE(geography_1, geography_2)
```

```googlesql
ST_MAKELINE(array_of_geography)
```

**Description**

Creates a `GEOGRAPHY` with a single linestring by
concatenating the point or line vertices of each of the input
`GEOGRAPHY`s in the order they are given.

`ST_MAKELINE` comes in two variants. For the first variant, input must be two
`GEOGRAPHY`s. For the second, input must be an `ARRAY` of type `GEOGRAPHY`. In
either variant, each input `GEOGRAPHY` must consist of one of the following
values:

+   Exactly one point.
+   Exactly one linestring.

For the first variant of `ST_MAKELINE`, if either input `GEOGRAPHY` is `NULL`,
`ST_MAKELINE` returns `NULL`. For the second variant, if input `ARRAY` or any
element in the input `ARRAY` is `NULL`, `ST_MAKELINE` returns `NULL`.

**Constraints**

Every edge must span strictly less than 180 degrees.

NOTE: The GoogleSQL snapping process may discard sufficiently short
edges and snap the two endpoints together. For instance, if two input
`GEOGRAPHY`s each contain a point and the two points are separated by a distance
less than the snap radius, the points will be snapped together. In such a case
the result will be a `GEOGRAPHY` with exactly one point.

**Return type**

LineString `GEOGRAPHY`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
