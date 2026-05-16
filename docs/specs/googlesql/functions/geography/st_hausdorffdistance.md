---
name: ST_HAUSDORFFDISTANCE
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  BindStHausdorffDistance computes
  `max over vertices(A) of (min point-to-segment distance to B)`
  using `s2.DistanceFromSegment`, then returns either the directed
  hAB value (when `directed => TRUE`) or the symmetric maximum of
  hAB and hBA. The previous vertex-only implementation overshot
  Example 1 by a factor of 2 because it never accounted for the
  closest point along an edge between two vertices of B.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_hausdorffdistance
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_hausdorffdistance.yaml
---

# ST_HAUSDORFFDISTANCE

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

## `ST_HAUSDORFFDISTANCE`

```googlesql
ST_HAUSDORFFDISTANCE(
  geography_1,
  geography_2
  [, directed => { TRUE | FALSE } ]
)
```

**Description**

Gets the discrete [Hausdorff distance][h-distance], which is the greatest of all
the distances from a discrete point in one geography to the closest
discrete point in another geography.

**Definitions**

+   `geography_1`: A `GEOGRAPHY` value that represents the first geography.
+   `geography_2`: A `GEOGRAPHY` value that represents the second geography.
+   `directed`: A named argument with a `BOOL` value. Represents the type of
    computation to use on the input geographies. If this argument isn't
    specified, `directed => FALSE` is used by default.

    +   `FALSE`: The largest Hausdorff distance found in
        (`geography_1`, `geography_2`) and
        (`geography_2`, `geography_1`).

    +   `TRUE` (default): The Hausdorff distance for
        (`geography_1`, `geography_2`).

**Details**

If an input geography is `NULL`, the function returns `NULL`.

**Return type**

`DOUBLE`

**Example**

The following query gets the Hausdorff distance between `geo1` and `geo2`:

```googlesql
WITH data AS (
  SELECT
    ST_GEOGFROMTEXT('LINESTRING(20 70, 70 60, 10 70, 70 70)') AS geo1,
    ST_GEOGFROMTEXT('LINESTRING(20 90, 30 90, 60 10, 90 10)') AS geo2
)
SELECT ST_HAUSDORFFDISTANCE(geo1, geo2, directed=>TRUE) AS distance
FROM data;

/*--------------------+
 | distance           |
 +--------------------+
 | 1688933.9832041925 |
 +--------------------*/
```

The following query gets the Hausdorff distance between `geo2` and `geo1`:

```googlesql
WITH data AS (
  SELECT
    ST_GEOGFROMTEXT('LINESTRING(20 70, 70 60, 10 70, 70 70)') AS geo1,
    ST_GEOGFROMTEXT('LINESTRING(20 90, 30 90, 60 10, 90 10)') AS geo2
)
SELECT ST_HAUSDORFFDISTANCE(geo2, geo1, directed=>TRUE) AS distance
FROM data;

/*--------------------+
 | distance           |
 +--------------------+
 | 5802892.745488612  |
 +--------------------*/
```

The following query gets the largest Hausdorff distance between
(`geo1` and `geo2`) and (`geo2` and `geo1`):

```googlesql
WITH data AS (
  SELECT
    ST_GEOGFROMTEXT('LINESTRING(20 70, 70 60, 10 70, 70 70)') AS geo1,
    ST_GEOGFROMTEXT('LINESTRING(20 90, 30 90, 60 10, 90 10)') AS geo2
)
SELECT ST_HAUSDORFFDISTANCE(geo1, geo2, directed=>FALSE) AS distance
FROM data;

/*--------------------+
 | distance           |
 +--------------------+
 | 5802892.745488612  |
 +--------------------*/
```

The following query produces the same results as the previous query because
`ST_HAUSDORFFDISTANCE` uses `directed=>FALSE` by default.

```googlesql
WITH data AS (
  SELECT
    ST_GEOGFROMTEXT('LINESTRING(20 70, 70 60, 10 70, 70 70)') AS geo1,
    ST_GEOGFROMTEXT('LINESTRING(20 90, 30 90, 60 10, 90 10)') AS geo2
)
SELECT ST_HAUSDORFFDISTANCE(geo1, geo2) AS distance
FROM data;
```

[h-distance]: http://en.wikipedia.org/wiki/Hausdorff_distance

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
