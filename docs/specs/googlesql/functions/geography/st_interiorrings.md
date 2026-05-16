---
name: ST_INTERIORRINGS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStInteriorRings now returns the inner rings as an
  ARRAY<GEOGRAPHY> of LineStrings (previously fell back to
  MULTILINESTRING). Each emitted ring is canonicalised — reversed to
  CCW winding, then rotated to start at the lex-smallest (lat, lng)
  vertex — to match BigQuery's display. A polygon without holes
  returns an empty array; `fullglobe` (now parsed as a distinct
  geography kind) likewise returns an empty array; only NULL input
  yields NULL.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_interiorrings
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_interiorrings.yaml
---

# ST_INTERIORRINGS

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

## `ST_INTERIORRINGS`

```googlesql
ST_INTERIORRINGS(polygon_geography)
```

**Description**

Returns an array of linestring geographies that corresponds to the interior
rings of a polygon geography. Each interior ring is the border of a hole within
the input polygon.

+   If the input geography is a polygon, excludes the outermost ring of the
    polygon geography and returns the linestrings corresponding to the interior
    rings.
+   If the input is the full `GEOGRAPHY`, returns an empty array.
+   If the input polygon has no holes, returns an empty array.
+   Returns an error if the input isn't a single polygon.

Use the `SAFE` prefix to return `NULL` for invalid input instead of an error.

**Return type**

`ARRAY<LineString GEOGRAPHY>`

**Examples**

```googlesql
WITH geo AS (
  SELECT ST_GEOGFROMTEXT('POLYGON((0 0, 1 1, 1 2, 0 0))') AS g UNION ALL
  SELECT ST_GEOGFROMTEXT('POLYGON((1 1, 1 10, 5 10, 5 1, 1 1), (2 2, 3 4, 2 4, 2 2))') UNION ALL
  SELECT ST_GEOGFROMTEXT('POLYGON((1 1, 1 10, 5 10, 5 1, 1 1), (2 2.5, 3.5 3, 2.5 2, 2 2.5), (3.5 7, 4 6, 3 3, 3.5 7))') UNION ALL
  SELECT ST_GEOGFROMTEXT('fullglobe') UNION ALL
  SELECT NULL)
SELECT ST_INTERIORRINGS(g) AS rings FROM geo;

/*----------------------------------------------------------------------------+
 | rings                                                                      |
 +----------------------------------------------------------------------------+
 | []                                                                         |
 | [LINESTRING(2 2, 3 4, 2 4, 2 2)]                                           |
 | [LINESTRING(2.5 2, 3.5 3, 2 2.5, 2.5 2), LINESTRING(3 3, 4 6, 3.5 7, 3 3)] |
 | []                                                                         |
 | NULL                                                                       |
 +----------------------------------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
