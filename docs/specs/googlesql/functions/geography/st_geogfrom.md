---
name: ST_GEOGFROM
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Dispatches to ST_GEOGFROMTEXT / ST_GEOGFROMGEOJSON / ST_GEOGFROMKML / ST_GEOGFROMWKB based on the input shape. Runtime entry: BindStGeogFrom in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geogfrom
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geogfrom.yaml
---

# ST_GEOGFROM

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

## `ST_GEOGFROM`

```googlesql
ST_GEOGFROM(expression)
```

**Description**

Converts an expression for a `STRING` or `BYTES` value into a
`GEOGRAPHY` value.

If `expression` represents a `STRING` value, it must be a valid
`GEOGRAPHY` representation in one of the following formats:

+ WKT format. To learn more about this format and the requirements to use it,
  see [ST_GEOGFROMTEXT][st-geogfromtext].
+ WKB in hexadecimal text format. To learn more about this format and the
  requirements to use it, see [ST_GEOGFROMWKB][st-geogfromwkb].
+ GeoJSON format. To learn more about this format and the
  requirements to use it, see [ST_GEOGFROMGEOJSON][st-geogfromgeojson].

If `expression` represents a `BYTES` value, it must be a valid `GEOGRAPHY`
binary expression in WKB format. To learn more about this format and the
requirements to use it, see [ST_GEOGFROMWKB][st-geogfromwkb].

If `expression` is `NULL`, the output is `NULL`.

**Return type**

`GEOGRAPHY`

**Examples**

This takes a WKT-formatted string and returns a `GEOGRAPHY` polygon:

```googlesql
SELECT ST_GEOGFROM('POLYGON((0 0, 0 2, 2 2, 2 0, 0 0))') AS WKT_format;

/*------------------------------------+
 | WKT_format                         |
 +------------------------------------+
 | POLYGON((2 0, 2 2, 0 2, 0 0, 2 0)) |
 +------------------------------------*/
```

This takes a WKB-formatted hexadecimal-encoded string and returns a
`GEOGRAPHY` point:

```googlesql
SELECT ST_GEOGFROM(FROM_HEX('010100000000000000000000400000000000001040')) AS WKB_format;

/*----------------+
 | WKB_format     |
 +----------------+
 | POINT(2 4)     |
 +----------------*/
```

This takes WKB-formatted bytes and returns a `GEOGRAPHY` point:

```googlesql
SELECT ST_GEOGFROM('010100000000000000000000400000000000001040') AS WKB_format;

/*----------------+
 | WKB_format     |
 +----------------+
 | POINT(2 4)     |
 +----------------*/
```

This takes a GeoJSON-formatted string and returns a `GEOGRAPHY` polygon:

```googlesql
SELECT ST_GEOGFROM(
  '{ "type": "Polygon", "coordinates": [ [ [2, 0], [2, 2], [1, 2], [0, 2], [0, 0], [2, 0] ] ] }'
) AS GEOJSON_format;

/*-----------------------------------------+
 | GEOJSON_format                          |
 +-----------------------------------------+
 | POLYGON((2 0, 2 2, 1 2, 0 2, 0 0, 2 0)) |
 +-----------------------------------------*/
```

[st-geogfromtext]: #st_geogfromtext

[st-geogfromwkb]: #st_geogfromwkb

[st-geogfromgeojson]: #st_geogfromgeojson

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
