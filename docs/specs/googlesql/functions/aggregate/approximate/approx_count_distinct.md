---
name: APPROX_COUNT_DISTINCT
dialect: googlesql
category: functions/aggregate/approximate
status: implemented
source_url: docs/third_party/googlesql-docs/approximate_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/approximate_aggregate_functions.md#approx_count_distinct
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/approximate/approx_count_distinct.yaml
---

# APPROX_COUNT_DISTINCT

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

## `APPROX_COUNT_DISTINCT`

```googlesql
APPROX_COUNT_DISTINCT(
  expression
  [ WHERE where_expression ]
)
```

**Description**

Returns the approximate result for `COUNT(DISTINCT expression)`. The value
returned is a statistical estimate, not necessarily the actual value.

This function is less accurate than `COUNT(DISTINCT expression)`, but performs
better on huge input.

**Supported Argument Types**

Any data type **except**:

+ `ARRAY`
+ `STRUCT`
+ `PROTO`

**Returned Data Types**

`INT64`

**Examples**

```googlesql
SELECT APPROX_COUNT_DISTINCT(x) as approx_distinct
FROM UNNEST([0, 1, 1, 2, 3, 5]) as x;

/*-----------------+
 | approx_distinct |
 +-----------------+
 | 5               |
 +-----------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/approximate_aggregate_functions.md`.
