---
name: ST_Y
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_y
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_y.yaml
---

# ST_Y

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

## `ST_Y`

```googlesql
ST_Y(point_geography_expression)
```

**Description**

Returns the latitude in degrees of the single-point input
`GEOGRAPHY`.

For any input `GEOGRAPHY` that isn't a single point,
including an empty `GEOGRAPHY`, `ST_Y` returns an
error. Use the `SAFE.` prefix to return `NULL` instead.

**Return type**

`DOUBLE`

**Example**

See [`ST_X`][st-x] for example usage.

[st-x]: #st_x

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
