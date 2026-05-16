---
name: COUNTIF
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#countif
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/countif.yaml
---

# COUNTIF

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

## `COUNTIF`

```googlesql
COUNTIF(
  [ DISTINCT ]
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
[ OVER over_clause ]

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  [ ORDER BY expression [ { ASC | DESC }  ] [, ...] ]
  [ window_frame_clause ]

```

**Description**

Gets the number of `TRUE` values for an expression.

**Definitions**

+ `expression`: A `BOOL` value that represents the expression to evaluate.
+   `DISTINCT`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `WHERE`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `HAVING { MAX | MIN }`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `OVER`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `over_clause`: To learn more, see
    [Aggregate function calls][aggregate-function-calls].
+   `window_specification`: To learn more, see
    [Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

[agg-threshold-clause]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#agg_threshold_clause

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Details**

The function signature `COUNTIF(DISTINCT ...)` is generally not useful. If you
would like to use `DISTINCT`, use `COUNT` with `DISTINCT IF`. For more
information, see the [`COUNT`][count] function.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT COUNTIF(x<0) AS num_negative, COUNTIF(x>0) AS num_positive
FROM UNNEST([5, -2, 3, 6, -10, -7, 4, 0]) AS x;

/*--------------+--------------+
 | num_negative | num_positive |
 +--------------+--------------+
 | 3            | 4            |
 +--------------+--------------*/
```

```googlesql
SELECT
  x,
  COUNTIF(x<0) OVER (ORDER BY ABS(x) ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING) AS num_negative
FROM UNNEST([5, -2, 3, 6, -10, NULL, -7, 4, 0]) AS x;

/*------+--------------+
 | x    | num_negative |
 +------+--------------+
 | NULL | 0            |
 | 0    | 1            |
 | -2   | 1            |
 | 3    | 1            |
 | 4    | 0            |
 | 5    | 0            |
 | 6    | 1            |
 | -7   | 2            |
 | -10  | 2            |
 +------+--------------*/
```

[count]: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#count

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
