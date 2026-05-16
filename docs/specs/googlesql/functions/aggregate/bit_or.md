---
name: BIT_OR
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#bit_or
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/bit_or.yaml
---

# BIT_OR

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

## `BIT_OR`

```googlesql
BIT_OR(
  [ DISTINCT ]
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
```

**Description**

Performs a bitwise OR operation on `expression` and returns the result.

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
SELECT BIT_OR(x) as bit_or FROM UNNEST([0xF001, 0x00A1]) as x;

/*--------+
 | bit_or |
 +--------+
 | 61601  |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
