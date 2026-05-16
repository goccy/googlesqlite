---
name: ST_NUMGEOMETRIES
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Number of constituent sub-geometries (1 for singletons, N for MULTI*). Runtime entry: BindStNumGeometries in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_numgeometries
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_numgeometries.yaml
---

# ST_NUMGEOMETRIES

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

## `ST_NUMGEOMETRIES`

```
ST_NUMGEOMETRIES(geography_expression)
```

**Description**

Returns the number of geometries in the input `GEOGRAPHY`. For a single point,
linestring, or polygon, `ST_NUMGEOMETRIES` returns `1`. For any collection of
geometries, `ST_NUMGEOMETRIES` returns the number of geometries making up the
collection. `ST_NUMGEOMETRIES` returns `0` if the input is the empty
`GEOGRAPHY`.

**Return type**

`INT64`

**Example**

The following example computes `ST_NUMGEOMETRIES` for a single point geography,
two collections, and an empty geography.

```googlesql
WITH example AS(
  SELECT ST_GEOGFROMTEXT('POINT(5 0)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('MULTIPOINT(0 1, 4 3, 2 6)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION EMPTY'))
SELECT
  geography,
  ST_NUMGEOMETRIES(geography) AS num_geometries,
FROM example;

/*------------------------------------------------------+----------------+
 | geography                                            | num_geometries |
 +------------------------------------------------------+----------------+
 | POINT(5 0)                                           | 1              |
 | MULTIPOINT(0 1, 4 3, 2 6)                            | 3              |
 | GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1)) | 2              |
 | GEOMETRYCOLLECTION EMPTY                             | 0              |
 +------------------------------------------------------+----------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
