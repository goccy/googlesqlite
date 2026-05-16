---
name: STDDEV_POP
dialect: googlesql
category: functions/aggregate/statistical
status: implemented
source_url: docs/third_party/googlesql-docs/statistical_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/statistical_aggregate_functions.md#stddev_pop
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/statistical/stddev_pop.yaml
---

# STDDEV_POP

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

## `STDDEV_POP`

```googlesql
STDDEV_POP(
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

Returns the population (biased) standard deviation of the values. The return
result is between `0` and `+Inf`.

All numeric types are supported. If the
input is `NUMERIC` or `BIGNUMERIC` then the internal aggregation is
stable with the final output converted to a `DOUBLE`.
Otherwise the input is converted to a `DOUBLE`
before aggregation, resulting in a potentially unstable result.

This function ignores any `NULL` inputs. If all inputs are ignored, this
function returns `NULL`. If this function receives a single non-`NULL` input,
it returns `0`.

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

`STDDEV_POP` can be used with differential privacy. To learn more, see
[Differentially private aggregate functions][dp-functions].

**Return Data Type**

`DOUBLE`

**Examples**

```googlesql
SELECT STDDEV_POP(x) AS results FROM UNNEST([10, 14, 18]) AS x

/*-------------------+
 | results           |
 +-------------------+
 | 3.265986323710904 |
 +-------------------*/
```

```googlesql
SELECT STDDEV_POP(x) AS results FROM UNNEST([10, 14, NULL]) AS x

/*---------+
 | results |
 +---------+
 | 2       |
 +---------*/
```

```googlesql
SELECT STDDEV_POP(x) AS results FROM UNNEST([10, NULL]) AS x

/*---------+
 | results |
 +---------+
 | 0       |
 +---------*/
```

```googlesql
SELECT STDDEV_POP(x) AS results FROM UNNEST([NULL]) AS x

/*---------+
 | results |
 +---------+
 | NULL    |
 +---------*/
```

```googlesql
SELECT STDDEV_POP(x) AS results FROM UNNEST([10, 14, CAST('Infinity' as DOUBLE)]) AS x

/*---------+
 | results |
 +---------+
 | NaN     |
 +---------*/
```

[dp-functions]: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`.
