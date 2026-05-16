---
name: IFERROR
dialect: googlesql
category: functions/debug
status: implemented
notes: |
  Runtime catches errors via a formatter-level safe-eval rewrite
  (every reachable function call routes through its
  `googlesqlite_safe_*` variant, ERROR(msg) folds to NULL, scalar
  subqueries get a cardinality guard). Example 6 —
  `IFERROR(IFERROR(ERROR('a'), ERROR('b')), 'c')` — closed by a
  narrow SQL pre-rewrite in
  internal/iferror_rewrite.go::applyIferrorTypePropagation. The
  rewrite wraps the inner IFERROR with
  `SAFE_CAST(... AS <T>)` where `<T>` is inferred from the outer
  catch arm when it's a string / numeric / bool literal, so the
  resolver's templated argument unification has the constraint it
  needs. The CAST is observationally a no-op because the inner
  `IFERROR(ERROR, ERROR)` always evaluates to NULL under the
  safe-eval rewrite, and `SAFE_CAST(NULL AS <T>)` stays NULL.
source_url: docs/third_party/googlesql-docs/debugging_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/debugging_functions.md#iferror
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/debug/iferror.yaml
---

# IFERROR

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

## `IFERROR`

```googlesql
IFERROR(try_expression, catch_expression)
```

**Description**

Evaluates `try_expression`.

When `try_expression` is evaluated:

+ If the evaluation of `try_expression` doesn't produce an error, then
  `IFERROR` returns the result of `try_expression` without evaluating
  `catch_expression`.
+ If the evaluation of `try_expression` produces a system error, then `IFERROR`
  produces that system error.
+ If the evaluation of `try_expression` produces an evaluation error, then
  `IFERROR` suppresses that evaluation error and evaluates `catch_expression`.

If `catch_expression` is evaluated:

+ If the evaluation of `catch_expression` doesn't produce an error, then
  `IFERROR` returns the result of `catch_expression`.
+ If the evaluation of `catch_expression` produces any error, then `IFERROR`
  produces that error.

**Arguments**

+ `try_expression`: An expression that returns a scalar value.
+ `catch_expression`: An expression that returns a scalar value.

The results of `try_expression` and `catch_expression` must share a
[supertype][supertype].

**Return Data Type**

The [supertype][supertype] for `try_expression` and
`catch_expression`.

**Example**

In the following example, the query successfully evaluates `try_expression`.

```googlesql
SELECT IFERROR('a', 'b') AS result

/*--------+
 | result |
 +--------+
 | a      |
 +--------*/
```

In the following example, the query successfully evaluates the
`try_expression` subquery.

```googlesql
SELECT IFERROR((SELECT [1,2,3][OFFSET(0)]), -1) AS result

/*--------+
 | result |
 +--------+
 | 1      |
 +--------*/
```

In the following example, `IFERROR` catches an evaluation error in the
`try_expression` and successfully evaluates `catch_expression`.

```googlesql
SELECT IFERROR(ERROR('a'), 'b') AS result

/*--------+
 | result |
 +--------+
 | b      |
 +--------*/
```

In the following example, `IFERROR` catches an evaluation error in the
`try_expression` subquery and successfully evaluates `catch_expression`.

```googlesql
SELECT IFERROR((SELECT [1,2,3][OFFSET(9)]), -1) AS result

/*--------+
 | result |
 +--------+
 | -1     |
 +--------*/
```

In the following query, the error is handled by the innermost `IFERROR`
operation, `IFERROR(ERROR('a'), 'b')`.

```googlesql
SELECT IFERROR(IFERROR(ERROR('a'), 'b'), 'c') AS result

/*--------+
 | result |
 +--------+
 | b      |
 +--------*/
```

In the following query, the error is handled by the outermost `IFERROR`
operation, `IFERROR(..., 'c')`.

```googlesql
SELECT IFERROR(IFERROR(ERROR('a'), ERROR('b')), 'c') AS result

/*--------+
 | result |
 +--------+
 | c      |
 +--------*/
```

In the following example, an evaluation error is produced because the subquery
passed in as the `try_expression` evaluates to a table, not a scalar value.

```googlesql
SELECT IFERROR((SELECT e FROM UNNEST([1, 2]) AS e), 3) AS result

/*--------+
 | result |
 +--------+
 | 3      |
 +--------*/
```

In the following example, `IFERROR` catches an evaluation error in `ERROR('a')`
and then evaluates `ERROR('b')`. Because there is also an evaluation error in
`ERROR('b')`, `IFERROR` produces an evaluation error for `ERROR('b')`.

```googlesql
SELECT IFERROR(ERROR('a'), ERROR('b')) AS result

--ERROR: OUT_OF_RANGE 'b'
```

[supertype]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md#supertypes

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/debugging_functions.md`.
