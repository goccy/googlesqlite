---
name: ST_DISJOINT
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Complement of ST_INTERSECTS. Runtime entry: BindStDisjoint in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_disjoint
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_disjoint.yaml
---

# ST_DISJOINT

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

## `ST_DISJOINT`

```googlesql
ST_DISJOINT(geography_1, geography_2)
```

**Description**

Returns `TRUE` if the intersection of `geography_1` and `geography_2` is empty,
that is, no point in `geography_1` also appears in `geography_2`.

`ST_DISJOINT` is the logical negation of [`ST_INTERSECTS`][st-intersects].

**Return type**

`BOOL`

[st-intersects]: #st_intersects

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
