---
name: ST_COVEREDBY
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Inverse of ST_COVERS. Runtime entry: BindStCoveredBy in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_coveredby
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_coveredby.yaml
---

# ST_COVEREDBY

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

## `ST_COVEREDBY`

```googlesql
ST_COVEREDBY(geography_1, geography_2)
```

**Description**

Returns `FALSE` if `geography_1` or `geography_2` is empty. Returns `TRUE` if no
points of `geography_1` lie in the exterior of `geography_2`.

Given two `GEOGRAPHY`s `a` and `b`,
`ST_COVEREDBY(a, b)` returns the same result as
[`ST_COVERS`][st-covers]`(b, a)`. Note the opposite order of arguments.

**Return type**

`BOOL`

[st-covers]: #st_covers

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
