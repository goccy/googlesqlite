---
name: ST_GEOGPOINTFROMGEOHASH
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Decodes a geohash to its centre POINT. Runtime entry: BindStGeogPointFromGeoHash in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geogpointfromgeohash
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geogpointfromgeohash.yaml
---

# ST_GEOGPOINTFROMGEOHASH

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

## `ST_GEOGPOINTFROMGEOHASH`

```googlesql
ST_GEOGPOINTFROMGEOHASH(geohash)
```

**Description**

Returns a `GEOGRAPHY` value that corresponds to a
point in the middle of a bounding box defined in the [GeoHash][geohash-link].

**Return type**

Point `GEOGRAPHY`

[geohash-link]: https://en.wikipedia.org/wiki/Geohash

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
