---
name: ARRAY_REVERSE
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array_reverse
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array_reverse.yaml
---

# ARRAY_REVERSE

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY_REVERSE`

```googlesql
ARRAY_REVERSE(value)
```

**Description**

Returns the input `ARRAY` with elements in reverse order.

**Return type**

`ARRAY`

**Examples**

```googlesql
SELECT ARRAY_REVERSE([1, 2, 3]) AS reverse_arr

/*-------------+
 | reverse_arr |
 +-------------+
 | [3, 2, 1]   |
 +-------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
