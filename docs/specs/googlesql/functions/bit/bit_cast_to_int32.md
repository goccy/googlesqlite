---
name: BIT_CAST_TO_INT32
dialect: googlesql
category: functions/bit
status: implemented
source_url: docs/third_party/googlesql-docs/bit_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/bit_functions.md#bit_cast_to_int32
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/bit/bit_cast_to_int32.yaml
---

# BIT_CAST_TO_INT32

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

## `BIT_CAST_TO_INT32`

```googlesql
BIT_CAST_TO_INT32(value)
```

**Description**

GoogleSQL supports bit casting to `INT32`. A bit
cast is a cast in which the order of bits is preserved instead of the value
those bytes represent.

The `value` parameter can represent:

+ `INT32`
+ `UINT32`

**Return Data Type**

`INT32`

**Examples**

```googlesql
SELECT BIT_CAST_TO_UINT32(-1) as UINT32_value, BIT_CAST_TO_INT32(BIT_CAST_TO_UINT32(-1)) as bit_cast_value;

/*---------------+----------------------+
 | UINT32_value  | bit_cast_value       |
 +---------------+----------------------+
 | 4294967295    | -1                   |
 +---------------+----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/bit_functions.md`.
