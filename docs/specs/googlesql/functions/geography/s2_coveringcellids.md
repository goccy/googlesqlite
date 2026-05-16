---
name: S2_COVERINGCELLIDS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindS2CoveringCellIDs now returns an ARRAY<INT64> directly instead
  of a JSON string. Defaults match upstream (min_level=0,
  max_level=30, max_cells=8); for points the result is the leaf
  cell clamped to [min_level, max_level] (truncated to maxLevel,
  dropped if its level falls below minLevel). POINT EMPTY yields an
  empty array. LINESTRING / POLYGON inputs run through
  `s2.RegionCoverer`. The auto-extracted testdata was restored from
  a broken row split (the docs-table cell wrapped across two lines).
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#s2_coveringcellids
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/s2_coveringcellids.yaml
---

# S2_COVERINGCELLIDS

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

## `S2_COVERINGCELLIDS`

```googlesql
S2_COVERINGCELLIDS(
    geography
    [, min_level => cell_level]
    [, max_level => cell_level]
    [, max_cells => max_cells]
    [, buffer => buffer])
```

**Description**

Returns an array of [S2 cell IDs][s2-cells-link] that cover the input
`GEOGRAPHY`. The function returns at most `max_cells` cells. The optional
arguments `min_level` and `max_level` specify minimum and maximum levels for
returned S2 cells. The array size is limited by the optional `max_cells`
argument. The optional `buffer` argument specifies a buffering factor in
meters; the region being covered is expanded from the extent of the
input geography by this amount.

This is advanced functionality for interoperability with systems utilizing the
[S2 Geometry Library][s2-root-link].

**Constraints**

+ Returns the cell ID as a signed `INT64` bit-equivalent to
  [unsigned 64-bit integer representation][s2-cells-link].
+ Can return negative cell IDs.
+ Valid S2 cell levels are 0 to 30.
+ `max_cells` defaults to 8 if not explicitly specified.
+ `buffer` should be nonnegative. It defaults to 0.0 meters if not explicitly
  specified.

**Return type**

`ARRAY<INT64>`

**Example**

```googlesql
WITH data AS (
  SELECT 1 AS id, ST_GEOGPOINT(-122, 47) AS geo
  UNION ALL
  SELECT 2 AS id, ST_GEOGFROMTEXT('POINT EMPTY') AS geo
  UNION ALL
  SELECT 3 AS id, ST_GEOGFROMTEXT('LINESTRING(-122.12 47.67, -122.19 47.69)') AS geo
)
SELECT id, S2_COVERINGCELLIDS(geo, min_level => 12) cells
FROM data;

/*----+--------------------------------------------------------------------------------------+
 | id | cells                                                                                |
 +----+--------------------------------------------------------------------------------------+
 | 1  | [6093613931972369317]                                                                |
 | 2  | []                                                                                   |
 | 3  | [6093384954555662336, 6093390709811838976, 6093390735581642752, 6093390740145045504, |
 |    |  6093390791416217600, 6093390812891054080, 6093390817187069952, 6093496378892222464] |
 +----+--------------------------------------------------------------------------------------*/
```

[s2-cells-link]: https://s2geometry.io/devguide/s2cell_hierarchy

[s2-root-link]: https://s2geometry.io/

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
