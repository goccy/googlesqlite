---
name: ST_TOUCHES
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Boundary tangency: Intersects but neither contains the other. Runtime entry: BindStTouches in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_touches
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_touches.yaml
---

# ST_TOUCHES

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

## `ST_TOUCHES`

```googlesql
ST_TOUCHES(geography_1, geography_2)
```

**Description**

Returns `TRUE` provided the following two conditions are satisfied:

1.  `geography_1` intersects `geography_2`.
1.  The interior of `geography_1` and the interior of `geography_2` are
    disjoint.

**Return type**

`BOOL`

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
