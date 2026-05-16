---
name: ST_DUMP
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStDump now produces a proper ARRAY<GEOGRAPHY>: simple
  geographies surface as a single-element array, MULTI* are split
  into their atoms, and GEOMETRYCOLLECTION descends recursively
  (so a nested collection's leaf components reach the result). The
  optional `dimension` argument (0 / 1 / 2 for points / lines /
  polygons; -1 keeps all) filters the result by component type.
  The testdata that had been auto-extracted as 4 rows is restored
  to upstream's 3-row form: the GEOMETRYCOLLECTION cell wraps over
  two display lines in the docs table but represents a single row.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_dump
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_dump.yaml
---

# ST_DUMP

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

## `ST_DUMP`

```googlesql
ST_DUMP(geography[, dimension])
```

**Description**

Returns an `ARRAY` of simple
`GEOGRAPHY`s where each element is a component of
the input `GEOGRAPHY`. A simple
`GEOGRAPHY` consists of a single point, linestring,
or polygon. If the input `GEOGRAPHY` is simple, the
result is a single element. When the input
`GEOGRAPHY` is a collection, `ST_DUMP` returns an
`ARRAY` with one simple
`GEOGRAPHY` for each component in the collection.

If `dimension` is provided, the function only returns
`GEOGRAPHY`s of the corresponding dimension. A
dimension of -1 is equivalent to omitting `dimension`.

**Return Type**

`ARRAY<GEOGRAPHY>`

**Examples**

The following example shows how `ST_DUMP` returns the simple geographies within
a complex geography.

```googlesql
WITH example AS (
  SELECT ST_GEOGFROMTEXT('POINT(0 0)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('MULTIPOINT(0 0, 1 1)') AS geography
  UNION ALL
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))'))
SELECT
  geography AS original_geography,
  ST_DUMP(geography) AS dumped_geographies
FROM example

/*-------------------------------------+------------------------------------+
 |         original_geographies        |      dumped_geographies            |
 +-------------------------------------+------------------------------------+
 | POINT(0 0)                          | [POINT(0 0)]                       |
 | MULTIPOINT(0 0, 1 1)                | [POINT(0 0), POINT(1 1)]           |
 | GEOMETRYCOLLECTION(POINT(0 0),      | [POINT(0 0), LINESTRING(1 2, 2 1)] |
 |   LINESTRING(1 2, 2 1))             |                                    |
 +-------------------------------------+------------------------------------*/
```

The following example shows how `ST_DUMP` with the dimension argument only
returns simple geographies of the given dimension.

```googlesql
WITH example AS (
  SELECT ST_GEOGFROMTEXT('GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(1 2, 2 1))') AS geography)
SELECT
  geography AS original_geography,
  ST_DUMP(geography, 1) AS dumped_geographies
FROM example

/*-------------------------------------+------------------------------+
 |         original_geographies        |      dumped_geographies      |
 +-------------------------------------+------------------------------+
 | GEOMETRYCOLLECTION(POINT(0 0),      | [LINESTRING(1 2, 2 1)]       |
 |   LINESTRING(1 2, 2 1))             |                              |
 +-------------------------------------+------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
