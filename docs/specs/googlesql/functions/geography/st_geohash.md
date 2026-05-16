---
name: ST_GEOHASH
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Base-32 geohash of a POINT at the given precision (default 12). Runtime entry: BindStGeoHash in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_geohash
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_geohash.yaml
---

# ST_GEOHASH

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

## `ST_GEOHASH`

```googlesql
ST_GEOHASH(geography_expression[, maxchars])
```

**Description**

Takes a single-point `GEOGRAPHY` and returns a [GeoHash][geohash-link]
representation of that `GEOGRAPHY` object.

+   `geography_expression`: Represents a `GEOGRAPHY` object. Only a `GEOGRAPHY`
    object that represents a single point is supported. If `ST_GEOHASH` is used
    over an empty `GEOGRAPHY` object, returns `NULL`.
+   `maxchars`: This optional `INT64` parameter specifies the maximum number of
    characters the hash will contain. Fewer characters corresponds to lower
    precision (or, described differently, to a bigger bounding box). `maxchars`
    defaults to 20 if not explicitly specified. A valid `maxchars` value is 1
    to 20. Any value below or above is considered unspecified and the default of
    20 is used.

**Return type**

`STRING`

**Example**

Returns a GeoHash of the Seattle Center with 10 characters of precision.

```googlesql
SELECT ST_GEOHASH(ST_GEOGPOINT(-122.35, 47.62), 10) geohash

/*--------------+
 | geohash      |
 +--------------+
 | c22yzugqw7   |
 +--------------*/
```

[geohash-link]: https://en.wikipedia.org/wiki/Geohash

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
