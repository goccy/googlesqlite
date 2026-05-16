---
name: ST_GEOGFROMGEOJSON
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Parses a GeoJSON document into a GEOGRAPHY. Runtime entry: BindStGeogFromGeoJSON in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geogfromgeojson
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geogfromgeojson.yaml
---

# ST_GEOGFROMGEOJSON

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

## `ST_GEOGFROMGEOJSON`

```googlesql
ST_GEOGFROMGEOJSON(
  geojson_string
  [, make_valid => constant_expression ]
)
```

**Description**

Returns a `GEOGRAPHY` value that corresponds to the
input [GeoJSON][geojson-link] representation.

`ST_GEOGFROMGEOJSON` accepts input that's [RFC 7946][geojson-spec-link]
compliant.

If the named argument `make_valid` is set to `TRUE`, the function attempts to
repair polygons that don't conform to [Open Geospatial Consortium][ogc-link]
semantics.

A GoogleSQL `GEOGRAPHY` has spherical
geodesic edges, whereas a GeoJSON `Geometry` object explicitly has planar edges.
To convert between these two types of edges, GoogleSQL adds additional
points to the line where necessary so that the resulting sequence of edges
remains within 10 meters of the original edge.

See [`ST_ASGEOJSON`][st-asgeojson] to format a
`GEOGRAPHY` as GeoJSON.

**Constraints**

The JSON input is subject to the following constraints:

+   `ST_GEOGFROMGEOJSON` only accepts JSON geometry fragments and can't be used
    to ingest a whole JSON document.
+   The input JSON fragment must consist of a GeoJSON geometry type, which
    includes `Point`, `MultiPoint`, `LineString`, `MultiLineString`, `Polygon`,
    `MultiPolygon`, and `GeometryCollection`. Any other GeoJSON type such as
    `Feature` or `FeatureCollection` will result in an error.
+   A position in the `coordinates` member of a GeoJSON geometry type must
    consist of exactly two elements. The first is the longitude and the second
    is the latitude. Therefore, `ST_GEOGFROMGEOJSON` doesn't support the
    optional third element for a position in the `coordinates` member.

**Return type**

`GEOGRAPHY`

[geojson-link]: https://en.wikipedia.org/wiki/GeoJSON

[geojson-spec-link]: https://tools.ietf.org/html/rfc7946

[ogc-link]: https://www.ogc.org/standards/sfa

[st-asgeojson]: #st_asgeojson

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
