---
name: VAR_SAMP
dialect: googlesql
category: functions/aggregate/statistical
status: implemented
source_url: docs/third_party/googlesql-docs/statistical_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/statistical_aggregate_functions.md#var_samp
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/statistical/var_samp.yaml
---

# VAR_SAMP

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

Verbatim copy from `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `VAR_SAMP`

```googlesql
VAR_SAMP(
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

Returns the sample (unbiased) variance of the values. The return result is
between `0` and `+Inf`.

All numeric types are supported. If the
input is `NUMERIC` or `BIGNUMERIC` then the internal aggregation is
stable with the final output converted to a `DOUBLE`.
Otherwise the input is converted to a `DOUBLE`
before aggregation, resulting in a potentially unstable result.

This function ignores any `NULL` inputs. If there are fewer than two non-`NULL`
inputs, this function returns `NULL`.

`NaN` is produced if:

+ Any input value is `NaN`
+ Any input value is positive infinity or negative infinity.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

This function can be used with the
[`AGGREGATION_THRESHOLD` clause][agg-threshold-clause].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

[agg-threshold-clause]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#agg_threshold_clause

<!-- mdlint on -->

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Return Data Type**

`DOUBLE`

**Examples**

```googlesql
SELECT VAR_SAMP(x) AS results FROM UNNEST([10, 14, 18]) AS x

/*---------+
 | results |
 +---------+
 | 16      |
 +---------*/
```

```googlesql
SELECT VAR_SAMP(x) AS results FROM UNNEST([10, 14, NULL]) AS x

/*---------+
 | results |
 +---------+
 | 8       |
 +---------*/
```

```googlesql
SELECT VAR_SAMP(x) AS results FROM UNNEST([10, NULL]) AS x

/*---------+
 | results |
 +---------+
 | NULL    |
 +---------*/
```

```googlesql
SELECT VAR_SAMP(x) AS results FROM UNNEST([NULL]) AS x

/*---------+
 | results |
 +---------+
 | NULL    |
 +---------*/
```

```googlesql
SELECT VAR_SAMP(x) AS results FROM UNNEST([10, 14, CAST('Infinity' as DOUBLE)]) AS x

/*---------+
 | results |
 +---------+
 | NaN     |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`.
