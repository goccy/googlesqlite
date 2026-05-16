---
name: BIT_XOR
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#bit_xor
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/bit_xor.yaml
---

# BIT_XOR

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

Verbatim copy from `docs/third_party/googlesql-docs/aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `BIT_XOR`

```googlesql
BIT_XOR(
  [ DISTINCT ]
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
```

**Description**

Performs a bitwise XOR operation on `expression` and returns the result.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

+ UINT32
+ UINT64
+ INT32
+ INT64

**Returned Data Types**

INT64

**Examples**

```googlesql
SELECT BIT_XOR(x) AS bit_xor FROM UNNEST([5678, 1234]) AS x;

/*---------+
 | bit_xor |
 +---------+
 | 4860    |
 +---------*/
```

```googlesql
SELECT BIT_XOR(x) AS bit_xor FROM UNNEST([1234, 5678, 1234]) AS x;

/*---------+
 | bit_xor |
 +---------+
 | 5678    |
 +---------*/
```

```googlesql
SELECT BIT_XOR(DISTINCT x) AS bit_xor FROM UNNEST([1234, 5678, 1234]) AS x;

/*---------+
 | bit_xor |
 +---------+
 | 4860    |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
