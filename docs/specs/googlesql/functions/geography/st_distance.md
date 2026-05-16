---
name: ST_DISTANCE
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_distance
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_distance.yaml
---

# ST_DISTANCE

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

## `ST_DISTANCE`

```
ST_DISTANCE(geography_1, geography_2[, use_spheroid])
```

**Description**

Returns the shortest distance in meters between two non-empty
`GEOGRAPHY`s.

If either of the input `GEOGRAPHY`s is empty,
`ST_DISTANCE` returns `NULL`.

The optional `use_spheroid` parameter determines how this function measures
distance. If `use_spheroid` is `FALSE`, the function measures distance on the
surface of a perfect sphere. If `use_spheroid` is `TRUE`, the function measures
distance on the surface of the [WGS84][wgs84-link] spheroid. The default value
of `use_spheroid` is `FALSE`.

**Return type**

`DOUBLE`

[wgs84-link]: https://en.wikipedia.org/wiki/World_Geodetic_System

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
