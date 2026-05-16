---
name: BIT_CAST_TO_UINT64
dialect: googlesql
category: functions/bit
status: implemented
source_url: docs/third_party/googlesql-docs/bit_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/bit_functions.md#bit_cast_to_uint64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/bit/bit_cast_to_uint64.yaml
---

# BIT_CAST_TO_UINT64

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

Verbatim copy from `docs/third_party/googlesql-docs/bit_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `BIT_CAST_TO_UINT64`

```googlesql
BIT_CAST_TO_UINT64(value)
```

**Description**

GoogleSQL supports bit casting to `UINT64`. A bit
cast is a cast in which the order of bits is preserved instead of the value
those bytes represent.

The `value` parameter can represent:

+ `INT64`
+ `UINT64`

**Return Data Type**

`UINT64`

**Example**

```googlesql
SELECT -1 as INT64_value, BIT_CAST_TO_UINT64(-1) as bit_cast_value;

/*--------------+----------------------+
 | INT64_value  | bit_cast_value       |
 +--------------+----------------------+
 | -1           | 18446744073709551615 |
 +--------------+----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/bit_functions.md`.
