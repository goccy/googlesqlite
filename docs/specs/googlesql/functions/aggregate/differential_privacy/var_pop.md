---
name: VAR_POP
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  Population variance with Laplace noise scaled by sensitivity²/epsilon.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#var_pop-differential_privacy
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/var_pop.yaml
---

# VAR_POP

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

## `VAR_POP` (`DIFFERENTIAL_PRIVACY`) 
<a id="dp_var_pop"></a>

```googlesql
WITH DIFFERENTIAL_PRIVACY ...
  VAR_POP(
    expression,
    [ contribution_bounds_per_row => (lower_bound, upper_bound) ]
  )
```

**Description**

Takes an expression and computes the population (biased) variance of the values
in the expression. The final result is an aggregation across
privacy unit columns between `0` and `+Inf`. You can
[clamp the input values][dp_clamped_named] explicitly, otherwise input values
are clamped implicitly. Clamping is performed per individual user values.

This function must be used with the `DIFFERENTIAL_PRIVACY` clause and
can support these arguments:

+ `expression`: The input expression. This can be any numeric input type,
  such as `INT64`. `NULL`s are always ignored.
+ `contribution_bounds_per_row`: A named argument with a
  [contribution bound][dp-clamped-named].
  Performs clamping for each row separately before performing intermediate
  grouping on individual user values.

`NUMERIC` and `BIGNUMERIC` arguments aren't allowed.
 If you need them, cast them as the
`DOUBLE` data type first.

**Return type**

`DOUBLE`

**Examples**

The following differentially private query gets the
population (biased) variance of items requested. Smaller aggregations may not
be included. This query references a view called
[`professors`][dp-example-tables].

```googlesql
-- With noise
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    VAR_POP(quantity, contribution_bounds_per_row => (0,100)) pop_variance
FROM professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations may be removed.
/*----------+-----------------+
 | item     | pop_variance    |
 +----------+-----------------+
 | pencil   | 642             |
 | pen      | 2.6666666666665 |
 | scissors | 2500            |
 +----------+-----------------*/
```

The following differentially private query gets the
population (biased) variance of items requested. Smaller aggregations might not
be included. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    VAR_POP(quantity, contribution_bounds_per_row=>(0, 100)) pop_variance
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+-----------------+
 | item     | pop_variance    |
 +----------+-----------------+
 | pencil   | 642             |
 | pen      | 2.6666666666665 |
 | scissors | 2500            |
 +----------+-----------------*/
```

[dp-example-tables]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_tables

[dp-clamped-named]: #dp_clamped_named

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
