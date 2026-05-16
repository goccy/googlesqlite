---
name: ST_WITHIN
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Inverse of ST_CONTAINS. Runtime entry: BindStWithin in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_within
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_within.yaml
---

# ST_WITHIN

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

## `ST_WITHIN`

```googlesql
ST_WITHIN(geography_1, geography_2)
```

**Description**

Returns `TRUE` if no point of `geography_1` is outside of `geography_2` and
the interiors of `geography_1` and `geography_2` intersect.

Given two geographies `a` and `b`, `ST_WITHIN(a, b)` returns the same result
as [`ST_CONTAINS`][st-contains]`(b, a)`. Note the opposite order of arguments.

**Return type**

`BOOL`

[st-contains]: #st_contains

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
