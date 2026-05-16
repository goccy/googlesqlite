---
name: ISERROR
dialect: googlesql
category: functions/debug
status: implemented
source_url: docs/third_party/googlesql-docs/debugging_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/debugging_functions.md#iserror
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/debug/iserror.yaml
---

# ISERROR

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

## `ISERROR`

```googlesql
ISERROR(try_expression)
```

**Description**

Evaluates `try_expression`.

+ If the evaluation of `try_expression` doesn't produce an error, then
  `ISERROR` returns `FALSE`.
+ If the evaluation of `try_expression` produces a system error, then `ISERROR`
  produces that system error.
+ If the evaluation of `try_expression` produces an evaluation error, then
  `ISERROR` returns `TRUE`.

**Arguments**

+ `try_expression`: An expression that returns a scalar value.

**Return Data Type**

`BOOL`

**Example**

In the following examples, `ISERROR` successfully evaluates `try_expression`.

```googlesql
SELECT ISERROR('a') AS is_error

/*----------+
 | is_error |
 +----------+
 | false    |
 +----------*/
```

```googlesql
SELECT ISERROR(2/1) AS is_error

/*----------+
 | is_error |
 +----------+
 | false    |
 +----------*/
```

```googlesql
SELECT ISERROR((SELECT [1,2,3][OFFSET(0)])) AS is_error

/*----------+
 | is_error |
 +----------+
 | false    |
 +----------*/
```

In the following examples, `ISERROR` catches an evaluation error in
`try_expression`.

```googlesql
SELECT ISERROR(ERROR('a')) AS is_error

/*----------+
 | is_error |
 +----------+
 | true     |
 +----------*/
```

```googlesql
SELECT ISERROR(2/0) AS is_error

/*----------+
 | is_error |
 +----------+
 | true     |
 +----------*/
```

```googlesql
SELECT ISERROR((SELECT [1,2,3][OFFSET(9)])) AS is_error

/*----------+
 | is_error |
 +----------+
 | true     |
 +----------*/
```

In the following example, an evaluation error is produced because the subquery
passed in as `try_expression` evaluates to a table, not a scalar value.

```googlesql
SELECT ISERROR((SELECT e FROM UNNEST([1, 2]) AS e)) AS is_error

/*----------+
 | is_error |
 +----------+
 | true     |
 +----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/debugging_functions.md`.
