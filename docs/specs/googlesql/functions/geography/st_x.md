---
name: ST_X
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_x
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_x.yaml
---

# ST_X

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

## `ST_X`

```googlesql
ST_X(point_geography_expression)
```

**Description**

Returns the longitude in degrees of the single-point input
`GEOGRAPHY`.

For any input `GEOGRAPHY` that isn't a single point,
including an empty `GEOGRAPHY`, `ST_X` returns an
error. Use the `SAFE.` prefix to obtain `NULL`.

**Return type**

`DOUBLE`

**Example**

The following example uses `ST_X` and `ST_Y` to extract coordinates from
single-point geographies.

```googlesql
WITH points AS
   (SELECT ST_GEOGPOINT(i, i + 1) AS p FROM UNNEST([0, 5, 12]) AS i)
 SELECT
   p,
   ST_X(p) as longitude,
   ST_Y(p) as latitude
FROM points;

/*--------------+-----------+----------+
 | p            | longitude | latitude |
 +--------------+-----------+----------+
 | POINT(0 1)   | 0.0       | 1.0      |
 | POINT(5 6)   | 5.0       | 6.0      |
 | POINT(12 13) | 12.0      | 13.0     |
 +--------------+-----------+----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
