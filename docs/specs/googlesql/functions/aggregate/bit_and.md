---
name: BIT_AND
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#bit_and
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/bit_and.yaml
---

# BIT_AND

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

## `BIT_AND`

```googlesql
BIT_AND(
  [ DISTINCT ]
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
```

**Description**

Performs a bitwise AND operation on `expression` and returns the result.

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
SELECT BIT_AND(x) as bit_and FROM UNNEST([0xF001, 0x00A1]) as x;

/*---------+
 | bit_and |
 +---------+
 | 1       |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
