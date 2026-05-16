---
name: BIT_CAST_TO_UINT32
dialect: googlesql
category: functions/conversion
status: implemented
source_url: docs/third_party/googlesql-docs/conversion_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/conversion_functions.md#bit_cast_to_uint32
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/conversion/bit_cast_to_uint32.yaml
---

# BIT_CAST_TO_UINT32

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

Verbatim copy from `docs/third_party/googlesql-docs/conversion_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `BIT_CAST_TO_UINT32`

```googlesql
BIT_CAST_TO_UINT32(value)
```

**Description**

GoogleSQL supports bit casting to `UINT32`. A bit
cast is a cast in which the order of bits is preserved instead of the value
those bytes represent.

The `value` parameter can represent:

+ `INT32`
+ `UINT32`

**Return Data Type**

`UINT32`

**Examples**

```googlesql
SELECT -1 as UINT32_value, BIT_CAST_TO_UINT32(-1) as bit_cast_value;

/*--------------+----------------------+
 | UINT32_value | bit_cast_value       |
 +--------------+----------------------+
 | -1           | 4294967295           |
 +--------------+----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/conversion_functions.md`.
