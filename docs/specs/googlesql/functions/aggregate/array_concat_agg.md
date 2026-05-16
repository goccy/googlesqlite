---
name: ARRAY_CONCAT_AGG
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#array_concat_agg
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/array_concat_agg.yaml
---

# ARRAY_CONCAT_AGG

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

## `ARRAY_CONCAT_AGG`

```googlesql
ARRAY_CONCAT_AGG(
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
  [ ORDER BY key [ { ASC | DESC } ] [, ... ] ]
  [ LIMIT n ]
)
```

**Description**

Concatenates elements from `expression` of type `ARRAY`, returning a single
array as a result.

This function ignores `NULL` input arrays, but respects the `NULL` elements in
non-`NULL` input arrays. Returns `NULL` if there are zero input rows or
`expression` evaluates to `NULL` for all rows.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

`ARRAY`

**Returned Data Types**

`ARRAY`

**Examples**

```googlesql
SELECT ARRAY_CONCAT_AGG(x) AS array_concat_agg FROM (
  SELECT [NULL, 1, 2, 3, 4] AS x
  UNION ALL SELECT NULL
  UNION ALL SELECT [5, 6]
  UNION ALL SELECT [7, 8, 9]
);

/*-----------------------------------+
 | array_concat_agg                  |
 +-----------------------------------+
 | [NULL, 1, 2, 3, 4, 5, 6, 7, 8, 9] |
 +-----------------------------------*/
```

```googlesql
SELECT ARRAY_CONCAT_AGG(x ORDER BY ARRAY_LENGTH(x)) AS array_concat_agg FROM (
  SELECT [1, 2, 3, 4] AS x
  UNION ALL SELECT [5, 6]
  UNION ALL SELECT [7, 8, 9]
);

/*-----------------------------------+
 | array_concat_agg                  |
 +-----------------------------------+
 | [5, 6, 7, 8, 9, 1, 2, 3, 4]       |
 +-----------------------------------*/
```

```googlesql
SELECT ARRAY_CONCAT_AGG(x LIMIT 2) AS array_concat_agg FROM (
  SELECT [1, 2, 3, 4] AS x
  UNION ALL SELECT [5, 6]
  UNION ALL SELECT [7, 8, 9]
);

/*--------------------------+
 | array_concat_agg         |
 +--------------------------+
 | [1, 2, 3, 4, 5, 6]       |
 +--------------------------*/
```

```googlesql
SELECT ARRAY_CONCAT_AGG(x ORDER BY ARRAY_LENGTH(x) LIMIT 2) AS array_concat_agg FROM (
  SELECT [1, 2, 3, 4] AS x
  UNION ALL SELECT [5, 6]
  UNION ALL SELECT [7, 8, 9]
);

/*------------------+
 | array_concat_agg |
 +------------------+
 | [5, 6, 7, 8, 9]  |
 +------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
