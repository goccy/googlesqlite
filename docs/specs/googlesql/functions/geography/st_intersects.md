---
name: ST_INTERSECTS
dialect: googlesql
category: functions/geography
status: implemented
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_intersects
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_intersects.yaml
---

# ST_INTERSECTS

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

## `ST_INTERSECTS`

```googlesql
ST_INTERSECTS(geography_1, geography_2)
```

**Description**

Returns `TRUE` if the point set intersection of `geography_1` and `geography_2`
is non-empty. Thus, this function returns `TRUE` if there is at least one point
that appears in both input `GEOGRAPHY`s.

If `ST_INTERSECTS` returns `TRUE`, it implies that [`ST_DISJOINT`][st-disjoint]
returns `FALSE`.

**Return type**

`BOOL`

[st-disjoint]: #st_disjoint

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
