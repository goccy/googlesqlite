---
name: APPROX_TOP_COUNT
dialect: googlesql
category: functions/aggregate/approximate
status: implemented
source_url: docs/third_party/googlesql-docs/approximate_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/approximate_aggregate_functions.md#approx_top_count
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/approximate/approx_top_count.yaml
---

# APPROX_TOP_COUNT

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

Verbatim copy from `docs/third_party/googlesql-docs/approximate_aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `APPROX_TOP_COUNT`

```googlesql
APPROX_TOP_COUNT(
  expression, number
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
```

**Description**

Returns the approximate top elements of `expression` as an array of `STRUCT`s.
The `number` parameter specifies the number of elements returned.

Each `STRUCT` contains two fields. The first field (named `value`) contains an
input value. The second field (named `count`) contains an `INT64` specifying the
number of times the value was returned.

Returns `NULL` if there are zero input rows.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

+ `expression`: Any data type that the `GROUP BY` clause supports.
+ `number`: `INT64` literal or query parameter.

**Returned Data Types**

`ARRAY<STRUCT>`

**Examples**

```googlesql
SELECT APPROX_TOP_COUNT(x, 2) as approx_top_count
FROM UNNEST(["apple", "apple", "pear", "pear", "pear", "banana"]) as x;

/*-------------------------+
 | approx_top_count        |
 +-------------------------+
 | [{pear, 3}, {apple, 2}] |
 +-------------------------*/
```

**NULL handling**

`APPROX_TOP_COUNT` doesn't ignore `NULL`s in the input. For example:

```googlesql
SELECT APPROX_TOP_COUNT(x, 2) as approx_top_count
FROM UNNEST([NULL, "pear", "pear", "pear", "apple", NULL]) as x;

/*------------------------+
 | approx_top_count       |
 +------------------------+
 | [{pear, 3}, {NULL, 2}] |
 +------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/approximate_aggregate_functions.md`.
