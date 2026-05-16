---
name: ST_ISRING
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  True when the input is a closed LINESTRING with no self-intersections. Runtime entry: BindStIsRing in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_isring
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_isring.yaml
---

# ST_ISRING

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

## `ST_ISRING`

```googlesql
ST_ISRING(geography_expression)
```

**Description**

Returns `TRUE` if the input `GEOGRAPHY` is a linestring and if the
linestring is both [`ST_ISCLOSED`][st-isclosed] and
simple. A linestring is considered simple if it doesn't pass through the
same point twice (with the exception of the start and endpoint, which may
overlap to form a ring).

An empty `GEOGRAPHY` isn't a ring.

**Return type**

`BOOL`

[st-isclosed]: #st_isclosed

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
