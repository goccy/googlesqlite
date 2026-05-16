---
name: ST_ASKML
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  KML fragment for POINT / LINESTRING / POLYGON / MULTI*. Runtime entry: BindStAsKML in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_askml
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_askml.yaml
---

# ST_ASKML

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

## `ST_ASKML`

```googlesql
ST_ASKML(geography)
```

**Description**

Takes a `GEOGRAPHY` and returns a `STRING` [KML geometry][kml-geometry-link].
Coordinates are formatted with as few digits as possible without loss
of precision.

**Return type**

`STRING`

[kml-geometry-link]: https://developers.google.com/kml/documentation/kmlreference#geometry

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
