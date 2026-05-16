---
name: ST_GEOGFROMWKB
dialect: googlesql
category: functions/geography
status: partial
notes: |
  BindStGeogFromWKB now accepts up to four arguments (the analyzer
  materialises named arguments — `planar`, `make_valid`, `oriented`
  — as positional after the leading WKB). STRING inputs are
  hex-decoded when they parse as hex; BYTES inputs are passed
  through verbatim. The WKT renderer's coordinate formatter folds
  ULP-level float noise (e.g. `0.9999999999999998`) down to its
  display-precision shape (`1`).
  
  Residual gap: BigQuery's `planar => TRUE` variant densifies each
  edge of the input line into intermediate vertices to approximate
  the geodesic on the sphere; our parser keeps the original vertex
  set unchanged, so Example 1's `planar` column shows
  `LINESTRING(1 1, 3 2)` instead of upstream's
  `LINESTRING(1 1, 2 1.5, 2.5 1.75, 3 2)`. Closing that requires a
  proper geodesic-aware subdivision rule. The geodesic column,
  POINT EMPTY handling, and all other cases match upstream.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geogfromwkb
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geogfromwkb.yaml
---

# ST_GEOGFROMWKB

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

## `ST_GEOGFROMWKB`

```googlesql
ST_GEOGFROMWKB(
  wkb_bytes_expression
  [ , oriented => value ]
  [ , planar => value ]
  [ , make_valid => value ]
)
```

```googlesql
ST_GEOGFROMWKB(
  wkb_hex_string_expression
  [, oriented => value ]
  [, planar => value ]
  [, make_valid => value ]
)
```

**Description**

Converts an expression from a hexadecimal-text `STRING` or `BYTES`
value into a `GEOGRAPHY` value. The expression must be in
[WKB][wkb-link] format.

To format `GEOGRAPHY` as WKB, use [`ST_ASBINARY`][st-asbinary].

**Definitions**

+   `wkb_bytes_expression`: A `BYTES` value that contains the [WKB][wkb-link]
    format.
+   `wkb_hex_string_expression`: A `STRING` value that contains the
    hexadecimal-encoded [WKB][wkb-link] format.
+   `oriented`: A named argument with a `BOOL` literal.

    +   If the value is `TRUE`, any polygons in the input are assumed to be
        oriented as follows: when traveling along the boundary of the polygon
        in the order of the input vertices, the interior of the polygon is on
        the left. This allows WKB to represent polygons larger than a
        hemisphere. See also [`ST_MAKEPOLYGONORIENTED`][st-makepolygonoriented],
        which is similar to `ST_GEOGFROMWKB` with `oriented=TRUE`.

    +   If the value is `FALSE` or omitted, this function returns the polygon
        with the smaller area.
+   `planar`: A named argument with a `BOOL` literal. If the value
    is `TRUE`, the edges of the linestrings and polygons are assumed to use
    planar map semantics, rather than GoogleSQL default spherical
    geodesics semantics.
+   `make_valid`: A named argument with a `BOOL` literal. If the
    value is `TRUE`, the function attempts to repair polygons that
    don't conform to [Open Geospatial Consortium][ogc-link] semantics.

**Details**

+   The function doesn't support three-dimensional geometries that have a `Z`
    suffix, nor does it support linear referencing system geometries with an `M`
    suffix.
+   `oriented` and `planar` can't be `TRUE` at the same time.
+   `oriented` and `make_valid` can't be `TRUE` at the same time.

**Return type**

`GEOGRAPHY`

**Example**

The following query reads the hex-encoded WKB data containing
`LINESTRING(1 1, 3 2)` and uses it with planar and geodesic semantics. When
planar is used, the function approximates the planar input line using
line that contains a chain of geodesic segments.

```googlesql
WITH wkb_data AS (
  SELECT '010200000002000000feffffffffffef3f000000000000f03f01000000000008400000000000000040' geo
)
SELECT
  ST_GeogFromWkb(geo, planar=>TRUE) AS from_planar,
  ST_GeogFromWkb(geo, planar=>FALSE) AS from_geodesic,
FROM wkb_data

/*---------------------------------------+----------------------+
 | from_planar                           | from_geodesic        |
 +---------------------------------------+----------------------+
 | LINESTRING(1 1, 2 1.5, 2.5 1.75, 3 2) | LINESTRING(1 1, 3 2) |
 +---------------------------------------+----------------------*/
```

[wkb-link]: https://en.wikipedia.org/wiki/Well-known_text#Well-known_binary

[st-asbinary]: #st_asbinary

[st-geogfromgeojson]: #st_geogfromgeojson

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
