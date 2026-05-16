---
name: ST_EXTERIORRING
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Outer ring of a POLYGON as a LINESTRING. Runtime entry: BindStExteriorRing in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_exteriorring
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_exteriorring.yaml
---

# ST_EXTERIORRING

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

## `ST_EXTERIORRING`

```googlesql
ST_EXTERIORRING(polygon_geography)
```

**Description**

Returns a linestring geography that corresponds to the outermost ring of a
polygon geography.

+   If the input geography is a polygon, gets the outermost ring of the polygon
    geography and returns the corresponding linestring.
+   If the input is the full `GEOGRAPHY`, returns an empty geography.
+   Returns an error if the input isn't a single polygon.

Use the `SAFE` prefix to return `NULL` for invalid input instead of an error.

**Return type**

+ Linestring `GEOGRAPHY`
+ Empty `GEOGRAPHY`

**Examples**

```googlesql
WITH geo as
 (SELECT ST_GEOGFROMTEXT('POLYGON((0 0, 1 4, 2 2, 0 0))') AS g UNION ALL
  SELECT ST_GEOGFROMTEXT('''POLYGON((1 1, 1 10, 5 10, 5 1, 1 1),
                                  (2 2, 3 4, 2 4, 2 2))''') as g)
SELECT ST_EXTERIORRING(g) AS ring FROM geo;

/*---------------------------------------+
 | ring                                  |
 +---------------------------------------+
 | LINESTRING(2 2, 1 4, 0 0, 2 2)        |
 | LINESTRING(5 1, 5 10, 1 10, 1 1, 5 1) |
 +---------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
