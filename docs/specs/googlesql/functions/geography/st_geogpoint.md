---
name: ST_GEOGPOINT
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geogpoint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geogpoint.yaml
---

# ST_GEOGPOINT

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

## `ST_GEOGPOINT`

```googlesql
ST_GEOGPOINT(longitude, latitude)
```

**Description**

Creates a `GEOGRAPHY` with a single point. `ST_GEOGPOINT` creates a point from
the specified `DOUBLE` longitude (in degrees,
negative west of the Prime Meridian, positive east) and latitude (in degrees,
positive north of the Equator, negative south) parameters and returns that point
in a `GEOGRAPHY` value.

NOTE: Some systems present latitude first; take care with argument order.

**Constraints**

+   Longitudes outside the range \[-180, 180\] are allowed; `ST_GEOGPOINT` uses
    the input longitude modulo 360 to obtain a longitude within \[-180, 180\].
+   Latitudes must be in the range \[-90, 90\]. Latitudes outside this range
    will result in an error.

**Return type**

Point `GEOGRAPHY`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
