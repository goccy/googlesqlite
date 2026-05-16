---
name: NULLIFERROR
dialect: googlesql
category: functions/debug
status: implemented
source_url: docs/third_party/googlesql-docs/debugging_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/debugging_functions.md#nulliferror
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/debug/nulliferror.yaml
---

# NULLIFERROR

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

Verbatim copy from `docs/third_party/googlesql-docs/debugging_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `NULLIFERROR`

```googlesql
NULLIFERROR(try_expression)
```

**Description**

Evaluates `try_expression`.

+ If the evaluation of `try_expression` doesn't produce an error, then
  `NULLIFERROR` returns the result of `try_expression`.
+ If the evaluation of `try_expression` produces a system error, then
 `NULLIFERROR` produces that system error.

+ If the evaluation of `try_expression` produces an evaluation error, then
  `NULLIFERROR` returns `NULL`.

**Arguments**

+ `try_expression`: An expression that returns a scalar value.

**Return Data Type**

The data type for `try_expression` or `NULL`

**Example**

In the following example, `NULLIFERROR` successfully evaluates
`try_expression`.

```googlesql
SELECT NULLIFERROR('a') AS result

/*--------+
 | result |
 +--------+
 | a      |
 +--------*/
```

In the following example, `NULLIFERROR` successfully evaluates
the `try_expression` subquery.

```googlesql
SELECT NULLIFERROR((SELECT [1,2,3][OFFSET(0)])) AS result

/*--------+
 | result |
 +--------+
 | 1      |
 +--------*/
```

In the following example, `NULLIFERROR` catches an evaluation error in
`try_expression`.

```googlesql
SELECT NULLIFERROR(ERROR('a')) AS result

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

In the following example, `NULLIFERROR` catches an evaluation error in
the `try_expression` subquery.

```googlesql
SELECT NULLIFERROR((SELECT [1,2,3][OFFSET(9)])) AS result

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

In the following example, an evaluation error is produced because the subquery
passed in as `try_expression` evaluates to a table, not a scalar value.

```googlesql
SELECT NULLIFERROR((SELECT e FROM UNNEST([1, 2]) AS e)) AS result

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/debugging_functions.md`.
