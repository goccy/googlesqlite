---
name: ST_CENTROID
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Area-weighted centroid for polygons / length-weighted for polylines / unweighted vertex centroid for points. Runtime entry: BindStCentroid in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_centroid
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_centroid.yaml
---

# ST_CENTROID

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

## `ST_CENTROID`

```googlesql
ST_CENTROID(geography_expression)
```

**Description**

Returns the _centroid_ of the input `GEOGRAPHY` as a single point `GEOGRAPHY`.

The _centroid_ of a `GEOGRAPHY` is the weighted average of the centroids of the
highest-dimensional components in the `GEOGRAPHY`. The centroid for components
in each dimension is defined as follows:

+   The centroid of points is the arithmetic mean of the input coordinates.
+   The centroid of linestrings is the centroid of all the edges weighted by
    length. The centroid of each edge is the geodesic midpoint of the edge.
+   The centroid of a polygon is its center of mass.

If the input `GEOGRAPHY` is empty, an empty `GEOGRAPHY` is returned.

**Constraints**

In the unlikely event that the centroid of a `GEOGRAPHY` can't be defined by a
single point on the surface of the Earth, a deterministic but otherwise
arbitrary point is returned. This can only happen if the centroid is exactly at
the center of the Earth, such as the centroid for a pair of antipodal points,
and the likelihood of this happening is vanishingly small.

**Return type**

Point `GEOGRAPHY`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
