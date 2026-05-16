---
name: BIT_CAST_TO_INT64
dialect: googlesql
category: functions/bit
status: implemented
source_url: docs/third_party/googlesql-docs/bit_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/bit_functions.md#bit_cast_to_int64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/bit/bit_cast_to_int64.yaml
---

# BIT_CAST_TO_INT64

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

## `BIT_CAST_TO_INT64`

```googlesql
BIT_CAST_TO_INT64(value)
```

**Description**

GoogleSQL supports bit casting to `INT64`. A bit
cast is a cast in which the order of bits is preserved instead of the value
those bytes represent.

The `value` parameter can represent:

+ `INT64`
+ `UINT64`

**Return Data Type**

`INT64`

**Example**

```googlesql
SELECT BIT_CAST_TO_UINT64(-1) as UINT64_value, BIT_CAST_TO_INT64(BIT_CAST_TO_UINT64(-1)) as bit_cast_value;

/*-----------------------+----------------------+
 | UINT64_value          | bit_cast_value       |
 +-----------------------+----------------------+
 | 18446744073709551615  | -1                   |
 +-----------------------+----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/bit_functions.md`.
