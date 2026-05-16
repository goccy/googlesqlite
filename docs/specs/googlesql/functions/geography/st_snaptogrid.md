---
name: ST_SNAPTOGRID
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Quantises every vertex to the given grid size. Runtime entry: BindStSnapToGrid in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_snaptogrid
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_snaptogrid.yaml
---

# ST_SNAPTOGRID

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

## `ST_SNAPTOGRID`

```googlesql
ST_SNAPTOGRID(geography_expression, grid_size)
```

**Description**

Returns the input `GEOGRAPHY`, where each vertex has
been snapped to a longitude/latitude grid. The grid size is determined by the
`grid_size` parameter which is given in degrees.

**Constraints**

Arbitrary grid sizes aren't supported. The `grid_size` parameter is rounded so
that it's of the form `10^n`, where `-10 < n < 0`.

**Return type**

`GEOGRAPHY`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
