---
name: PERCENTILE_CONT
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  Same approximation as ANON_PERCENTILE_CONT.
  
  Implementation lives in internal/functions/aggregate/dp.go;
  the analyzer's Anonymization rewriter is enabled in
  internal/analyzer.go and lowers SELECT WITH DIFFERENTIAL_PRIVACY
  / ANONYMIZATION queries into $differential_privacy_<op> /
  $anon_<op> aggregate calls. The DifferentialPrivacyAggregateScan
  handler in internal/formatter.go captures the per-scan
  (epsilon, delta) options and appends them as runtime args to
  every DP aggregate call. Catalog registration of the internal
  $-prefixed signatures uses BuiltinFunctionOptions.IncludeFunctionIds
  (see newDPBuiltinFunctionOptions in internal/catalog.go).
source_url: docs/third_party/googlesql-docs/aggregate-dp-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#percentile_cont-differential_privacy
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/percentile_cont.yaml
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

Verbatim copy from `docs/third_party/googlesql-docs/aggregate-dp-functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PERCENTILE_CONT` (`DIFFERENTIAL_PRIVACY`) 
<a id="dp_percentile_cont"></a>

```googlesql
WITH DIFFERENTIAL_PRIVACY ...
  PERCENTILE_CONT(
    expression,
    percentile,
    contribution_bounds_per_row => (lower_bound, upper_bound)
  )
```

**Description**

Takes an expression and computes a percentile for it. The final result is an
aggregation across privacy unit columns.

This function must be used with the [`DIFFERENTIAL_PRIVACY` clause][dp-syntax]
and can support these arguments:

+ `expression`: The input expression. This can be most numeric input types,
  such as `INT64`. `NULL` values are always ignored.
+ `percentile`: The percentile to compute. The percentile must be a literal in
  the range `[0, 1]`.
+ `contribution_bounds_per_row`: A named argument with a
  [contribution bounds][dp-clamped-named].
  Performs clamping for each row separately before performing intermediate
  grouping on the privacy unit column.

`NUMERIC` and `BIGNUMERIC` arguments aren't allowed.
 If you need them, cast them as the
`DOUBLE` data type first.

**Return type**

`DOUBLE`

**Examples**

The following differentially private query gets the percentile of items
requested. Smaller aggregations might not be included. This query references a
view called [`professors`][dp-example-tables].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    PERCENTILE_CONT(quantity, 0.5, contribution_bounds_per_row => (0,100)) percentile_requested
FROM professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
 /*----------+----------------------+
  | item     | percentile_requested |
  +----------+----------------------+
  | pencil   | 72.00011444091797    |
  | scissors | 8.000175476074219    |
  | pen      | 23.001075744628906   |
  +----------+----------------------*/
```

The following differentially private query gets the percentile of items
requested. Smaller aggregations might not be included. This query references a
view called [`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    PERCENTILE_CONT(quantity, 0.5, contribution_bounds_per_row=>(0, 100)) percentile_requested
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+----------------------+
 | item     | percentile_requested |
 +----------+----------------------+
 | pencil   | 72.00011444091797    |
 | scissors | 8.000175476074219    |
 | pen      | 23.001075744628906   |
 +----------+----------------------*/
```

[dp-example-tables]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_tables

[dp-clamped-named]: #dp_clamped_named

[dp-syntax]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_clause

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
