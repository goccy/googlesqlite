---
name: ST_ASGEOJSON
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Renders the geography as a GeoJSON document (Point / LineString / Polygon / Multi* shapes). Runtime entry: BindStAsGeoJSON in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_asgeojson
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_asgeojson.yaml
---

# ST_ASGEOJSON

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

## `ST_ASGEOJSON`

```googlesql
ST_ASGEOJSON(geography_expression)
```

**Description**

Returns the [RFC 7946][GeoJSON-spec-link] compliant [GeoJSON][geojson-link]
representation of the input `GEOGRAPHY`.

A GoogleSQL `GEOGRAPHY` has spherical
geodesic edges, whereas a GeoJSON `Geometry` object explicitly has planar edges.
To convert between these two types of edges, GoogleSQL adds additional
points to the line where necessary so that the resulting sequence of edges
remains within 10 meters of the original edge.

See [`ST_GEOGFROMGEOJSON`][st-geogfromgeojson] to construct a
`GEOGRAPHY` from GeoJSON.

**Return type**

`STRING`

[geojson-spec-link]: https://tools.ietf.org/html/rfc7946

[geojson-link]: https://en.wikipedia.org/wiki/GeoJSON

[st-geogfromgeojson]: #st_geogfromgeojson

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
