---
name: PERCENTILE_CONT
dialect: googlesql
category: functions/window
status: implemented
source_url: docs/third_party/googlesql-docs/navigation_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/navigation_functions.md#percentile_cont
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/window/percentile_cont.yaml
---

# PERCENTILE_CONT

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

Verbatim copy from `docs/third_party/googlesql-docs/navigation_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PERCENTILE_CONT`

```googlesql
PERCENTILE_CONT (value_expression, percentile [{RESPECT | IGNORE} NULLS])
OVER over_clause

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]

```

**Description**

Computes the specified percentile value for the value_expression, with linear
interpolation.

This function ignores NULL
values if
`RESPECT NULLS` is absent. If `RESPECT NULLS` is present:

+ Interpolation between two `NULL` values returns `NULL`.
+ Interpolation between a `NULL` value and a non-`NULL` value returns the
  non-`NULL` value.

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

`PERCENTILE_CONT` can be used with differential privacy. To learn more, see
[Differentially private aggregate functions][dp-functions].

**Supported Argument Types**

+ `value_expression` and `percentile` must have one of the following types:
   + `NUMERIC`
   + `BIGNUMERIC`
   + `DOUBLE`
+ `percentile` must be a literal in the range `[0, 1]`.

**Return Data Type**

The return data type is determined by the argument types with the following
table.
<table>

<thead>
<tr>
<th>INPUT</th><th><code>NUMERIC</code></th><th><code>BIGNUMERIC</code></th><th><code>DOUBLE</code></th>
</tr>
</thead>
<tbody>
<tr><th><code>NUMERIC</code></th><td style="vertical-align:middle"><code>NUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>BIGNUMERIC</code></th><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>BIGNUMERIC</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
<tr><th><code>DOUBLE</code></th><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td><td style="vertical-align:middle"><code>DOUBLE</code></td></tr>
</tbody>

</table>

**Examples**

The following example computes the value for some percentiles from a column of
values while ignoring nulls.

```googlesql
SELECT
  PERCENTILE_CONT(x, 0) OVER() AS min,
  PERCENTILE_CONT(x, 0.01) OVER() AS percentile1,
  PERCENTILE_CONT(x, 0.5) OVER() AS median,
  PERCENTILE_CONT(x, 0.9) OVER() AS percentile90,
  PERCENTILE_CONT(x, 1) OVER() AS max
FROM UNNEST([0, 3, NULL, 1, 2]) AS x LIMIT 1;

 /*-----+-------------+--------+--------------+-----+
  | min | percentile1 | median | percentile90 | max |
  +-----+-------------+--------+--------------+-----+
  | 0   | 0.03        | 1.5    | 2.7          | 3   |
  +-----+-------------+--------+--------------+-----*/
```

The following example computes the value for some percentiles from a column of
values while respecting nulls.

```googlesql
SELECT
  PERCENTILE_CONT(x, 0 RESPECT NULLS) OVER() AS min,
  PERCENTILE_CONT(x, 0.01 RESPECT NULLS) OVER() AS percentile1,
  PERCENTILE_CONT(x, 0.5 RESPECT NULLS) OVER() AS median,
  PERCENTILE_CONT(x, 0.9 RESPECT NULLS) OVER() AS percentile90,
  PERCENTILE_CONT(x, 1 RESPECT NULLS) OVER() AS max
FROM UNNEST([0, 3, NULL, 1, 2]) AS x LIMIT 1;

/*------+-------------+--------+--------------+-----+
 | min  | percentile1 | median | percentile90 | max |
 +------+-------------+--------+--------------+-----+
 | NULL | 0           | 1      | 2.6          | 3   |
 +------+-------------+--------+--------------+-----*/
```

[dp-functions]: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/navigation_functions.md`.
