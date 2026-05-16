---
name: ST_GEOMETRYTYPE
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  ST_Point / ST_LineString / ST_Polygon / ST_MultiPoint / ST_MultiLineString / ST_MultiPolygon strings. Runtime entry: BindStGeometryType in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geometrytype
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geometrytype.yaml
---

# ST_GEOMETRYTYPE

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

## `ST_GEOMETRYTYPE`

```googlesql
ST_GEOMETRYTYPE(geography_expression)
```

**Description**

Returns the [Open Geospatial Consortium][ogc-link] (OGC) geometry type that
describes the input `GEOGRAPHY`. The OGC geometry type matches the
types that are used in [WKT][wkt-link] and [GeoJSON][geojson-link] formats and
printed for [ST_ASTEXT][st-astext] and [ST_ASGEOJSON][st-asgeojson].
`ST_GEOMETRYTYPE` returns the OGC geometry type with the "ST_" prefix.

`ST_GEOMETRYTYPE` returns the following given the type on the input:

+   Single point geography: Returns `ST_Point`.
+   Collection of only points: Returns `ST_MultiPoint`.
+   Single linestring geography: Returns `ST_LineString`.
+   Collection of only linestrings: Returns `ST_MultiLineString`.
+   Single polygon geography: Returns `ST_Polygon`.
+   Collection of only polygons: Returns `ST_MultiPolygon`.
+   Collection with elements of different dimensions, or the input is the empty
    geography: Returns `ST_GeometryCollection`.

**Return type**

`STRING`

**Example**

The following example shows how `ST_GEOMETRYTYPE` takes geographies and returns
the names of their OGC geometry types.

```googlesql
WITH example AS(
  SELECT ST_GEOGFROMTEXT('POINT(0 1)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('MULTILINESTRING((2 2, 3 4), (5 6, 7 7))')
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(MULTIPOINT(-1 2, 0 12), LINESTRING(-2 4, 0 6))')
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION EMPTY'))
SELECT
  geography AS WKT,
  ST_GEOMETRYTYPE(geography) AS geometry_type_name
FROM example;

/*-------------------------------------------------------------------+-----------------------+
 | WKT                                                               | geometry_type_name    |
 +-------------------------------------------------------------------+-----------------------+
 | POINT(0 1)                                                        | ST_Point              |
 | MULTILINESTRING((2 2, 3 4), (5 6, 7 7))                           | ST_MultiLineString    |
 | GEOMETRYCOLLECTION(MULTIPOINT(-1 2, 0 12), LINESTRING(-2 4, 0 6)) | ST_GeometryCollection |
 | GEOMETRYCOLLECTION EMPTY                                          | ST_GeometryCollection |
 +-------------------------------------------------------------------+-----------------------*/
```

[ogc-link]: https://www.ogc.org/standards/sfa

[wkt-link]: https://en.wikipedia.org/wiki/Well-known_text

[geojson-link]: https://en.wikipedia.org/wiki/GeoJSON

[st-astext]: #st_astext

[st-asgeojson]: #st_asgeojson

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
