---
name: ST_NPOINTS
dialect: googlesql
category: functions/geography
status: implemented
notes: |
  Total point count (alias of ST_NUMPOINTS). Runtime entry: BindStNPoints in internal/functions/geography/.
source_url: docs/third_party/googlesql-docs/geography_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/geography_functions.md#st_npoints
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/geography/st_npoints.yaml
---

# ST_NPOINTS

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

## `ST_NPOINTS`

```googlesql
ST_NPOINTS(geography_expression)
```

**Description**

An alias of [ST_NUMPOINTS][st-numpoints].

[st-numpoints]: #st_numpoints

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/geography_functions.md`.
