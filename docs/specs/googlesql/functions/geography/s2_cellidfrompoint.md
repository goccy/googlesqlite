---
name: S2_CELLIDFROMPOINT
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  S2 cell-id covering the input POINT at the requested level (default 30). Runtime entry: BindS2CellIDFromPoint in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#s2_cellidfrompoint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/s2_cellidfrompoint.yaml
---

# S2_CELLIDFROMPOINT

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

## `S2_CELLIDFROMPOINT`

```googlesql
S2_CELLIDFROMPOINT(point_geography[, level => cell_level])
```

**Description**

Returns the [S2 cell ID][s2-cells-link] covering a point `GEOGRAPHY`.

+ The optional `INT64` parameter `level` specifies the S2 cell level for the
  returned cell. Naming this argument is optional.

This is advanced functionality for interoperability with systems utilizing the
[S2 Geometry Library][s2-root-link].

**Constraints**

+ Returns the cell ID as a signed `INT64` bit-equivalent to
  [unsigned 64-bit integer representation][s2-cells-link].
+ Can return negative cell IDs.
+ Valid S2 cell levels are 0 to 30.
+ `level` defaults to 30 if not explicitly specified.
+ The function only supports a single point GEOGRAPHY. Use the `SAFE` prefix if
  the input can be multipoint, linestring, polygon, or an empty `GEOGRAPHY`.
+ To compute the covering of a complex `GEOGRAPHY`, use
  [S2_COVERINGCELLIDS][s2-coveringcellids].

**Return type**

`INT64`

**Example**

```googlesql
WITH data AS (
  SELECT 1 AS id, ST_GEOGPOINT(-122, 47) AS geo
  UNION ALL
  -- empty geography isn't supported
  SELECT 2 AS id, ST_GEOGFROMTEXT('POINT EMPTY') AS geo
  UNION ALL
  -- only points are supported
  SELECT 3 AS id, ST_GEOGFROMTEXT('LINESTRING(1 2, 3 4)') AS geo
)
SELECT id,
       SAFE.S2_CELLIDFROMPOINT(geo) cell30,
       SAFE.S2_CELLIDFROMPOINT(geo, level => 10) cell10
FROM data;

/*----+---------------------+---------------------+
 | id | cell30              | cell10              |
 +----+---------------------+---------------------+
 | 1  | 6093613931972369317 | 6093613287902019584 |
 | 2  | NULL                | NULL                |
 | 3  | NULL                | NULL                |
 +----+---------------------+---------------------*/
```

[s2-cells-link]: https://s2geometry.io/devguide/s2cell_hierarchy

[s2-root-link]: https://s2geometry.io/

[s2-coveringcellids]: #s2_coveringcellids

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
