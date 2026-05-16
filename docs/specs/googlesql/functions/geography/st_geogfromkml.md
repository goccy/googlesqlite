---
name: ST_GEOGFROMKML
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Parses a small subset of KML (Point / LineString / Polygon) into a GEOGRAPHY. Runtime entry: BindStGeogFromKML in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geogfromkml
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geogfromkml.yaml
---

# ST_GEOGFROMKML

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

## `ST_GEOGFROMKML`

```googlesql
ST_GEOGFROMKML(kml_geometry)
```

Takes a `STRING` [KML geometry][kml-geometry-link] and returns a
`GEOGRAPHY`. The KML geomentry can include:

+  Point with coordinates element only
+  Linestring with coordinates element only
+  Polygon with boundary elements only
+  Multigeometry

[kml-geometry-link]: https://developers.google.com/kml/documentation/kmlreference#geometry

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
