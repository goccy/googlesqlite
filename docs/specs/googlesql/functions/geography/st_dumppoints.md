---
name: ST_DUMPPOINTS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStDumpPoints returns every vertex of the geography as an
  ARRAY<GEOGRAPHY> of POINT atoms (previously emitted as a
  MULTIPOINT, which the analyzer's return type couldn't accept).
  GEOMETRYCOLLECTION descends into its parts. Testdata restored to
  upstream's 3-row form (the auto-extraction had split the wrapped
  GEOMETRYCOLLECTION cell into 4 rows).
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_dumppoints
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_dumppoints.yaml
---

# ST_DUMPPOINTS

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

## `ST_DUMPPOINTS`

```googlesql
ST_DUMPPOINTS(geography)
```

**Description**

Takes an input geography and returns all of its points, line vertices, and
polygon vertices as an array of point geographies.

**Return Type**

`ARRAY<Point GEOGRAPHY>`

**Examples**

```googlesql
WITH example AS (
  SELECT ST_GEOGFROMTEXT('POINT(0 0)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('MULTIPOINT(0 0, 1 1)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))'))
SELECT
  geography AS original_geography,
  ST_DUMPPOINTS(geography) AS dumped_points_geographies
FROM example

/*-------------------------------------+------------------------------------+
 | original_geographies                | dumped_points_geographies          |
 +-------------------------------------+------------------------------------+
 | POINT(0 0)                          | [POINT(0 0)]                       |
 | MULTIPOINT(0 0, 1 1)                | [POINT(0 0),POINT(1 1)]            |
 | GEOMETRYCOLLECTION(POINT(0 0),      | [POINT(0 0),POINT(1 2),POINT(2 1)] |
 |   LINESTRING(1 2, 2 1))             |                                    |
 +-------------------------------------+------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
