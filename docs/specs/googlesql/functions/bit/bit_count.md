---
name: BIT_COUNT
dialect: googlesql
category: functions/bit
status: implemented
source_url: docs/third_party/googlesql-docs/bit_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/bit_functions.md#bit_count
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/bit/bit_count.yaml
---

# BIT_COUNT

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

## `BIT_COUNT`

```googlesql
BIT_COUNT(expression)
```

**Description**

The input, `expression`, must be an
integer or `BYTES`.

Returns the number of bits that are set in the input `expression`.
For signed integers, this is the number of bits in two's complement form.

**Return Data Type**

`INT64`

**Example**

```googlesql
SELECT a, BIT_COUNT(a) AS a_bits, FORMAT("%T", b) as b, BIT_COUNT(b) AS b_bits
FROM UNNEST([
  STRUCT(0 AS a, b'' AS b), (0, b'\x00'), (5, b'\x05'), (8, b'\x00\x08'),
  (0xFFFF, b'\xFF\xFF'), (-2, b'\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFE'),
  (-1, b'\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF'),
  (NULL, b'\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF')
]) AS x;

/*-------+--------+---------------------------------------------+--------+
 | a     | a_bits | b                                           | b_bits |
 +-------+--------+---------------------------------------------+--------+
 | 0     | 0      | b""                                         | 0      |
 | 0     | 0      | b"\x00"                                     | 0      |
 | 5     | 2      | b"\x05"                                     | 2      |
 | 8     | 1      | b"\x00\x08"                                 | 1      |
 | 65535 | 16     | b"\xff\xff"                                 | 16     |
 | -2    | 63     | b"\xff\xff\xff\xff\xff\xff\xff\xfe"         | 63     |
 | -1    | 64     | b"\xff\xff\xff\xff\xff\xff\xff\xff"         | 64     |
 | NULL  | NULL   | b"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" | 80     |
 +-------+--------+---------------------------------------------+--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/bit_functions.md`.
