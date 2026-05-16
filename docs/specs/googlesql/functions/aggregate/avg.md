---
name: AVG
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#avg
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/avg.yaml
---

# AVG

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

## `AVG`

```googlesql
AVG(
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

Returns the average of non-`NULL` values in an aggregated group.

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

`AVG` can be used with differential privacy. For more information, see
[Differentially private aggregate functions][dp-functions].

Caveats:

+ If the aggregated group is empty or the argument is `NULL` for all rows in
  the group, returns `NULL`.
+ If the argument is `NaN` for any row in the group, returns `NaN`.
+ If the argument is `[+|-]Infinity` for any row in the group, returns either
  `[+|-]Infinity` or `NaN`.
+ If there is numeric overflow, produces an error.
+ If a [floating-point type][floating-point-types] is returned, the result is
  [non-deterministic][non-deterministic], which means you might receive a
  different result each time you use this function.

[floating-point-types]: https://github.com/google/googlesql/blob/master/docs/data-types.md#floating_point_types

[non-deterministic]: https://github.com/google/googlesql/blob/master/docs/data-types.md#floating_point_semantics

**Supported Argument Types**

+ Any numeric input type
+ `INTERVAL`

**Returned Data Types**

<table>

<thead>
<tr>
<th>INPUT</th><th><code>INT32</code></th><th><code>INT64</code></th><th><code>UINT32</code></th><th><code>UINT64</code></th><th><code>NUMERIC</code></th><th><code>BIGNUMERIC</code></th><th><code>FLOAT</code></th><th><code>DOUBLE</code></th><th><code>INTERVAL</code></th>
</tr>
</thead>
<tbody>
<tr><th>OUTPUT</th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>INTERVAL</code></td></tr>
</tbody>

</table>

**Examples**

```googlesql
SELECT AVG(x) as avg
FROM UNNEST([0, 2, 4, 4, 5]) as x;

/*-----+
 | avg |
 +-----+
 | 3   |
 +-----*/
```

```googlesql
SELECT AVG(DISTINCT x) AS avg
FROM UNNEST([0, 2, 4, 4, 5]) AS x;

/*------+
 | avg  |
 +------+
 | 2.75 |
 +------*/
```

```googlesql
SELECT
  x,
  AVG(x) OVER (ORDER BY x ROWS BETWEEN 1 PRECEDING AND CURRENT ROW) AS avg
FROM UNNEST([0, 2, NULL, 4, 4, 5]) AS x;

/*------+------+
 | x    | avg  |
 +------+------+
 | NULL | NULL |
 | 0    | 0    |
 | 2    | 1    |
 | 4    | 3    |
 | 4    | 4    |
 | 5    | 4.5  |
 +------+------*/
```

[dp-functions]: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
