---
name: SUM
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  DifferentialPrivacyAggregateScan path: clamps contributions to (lo, hi), sums them, adds Laplace noise with scale = max(|lo|, |hi|) / epsilon.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#sum-differential_privacy
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/sum.yaml
---

# SUM

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

## `SUM` (`DIFFERENTIAL_PRIVACY`) 
<a id="dp_sum"></a>

```googlesql
WITH DIFFERENTIAL_PRIVACY ...
  SUM(
    expression,
    [ contribution_bounds_per_group => (lower_bound, upper_bound) ]
  )
```

**Description**

Returns the sum of non-`NULL`, non-`NaN` values in the expression. The final
result is an aggregation across privacy unit columns.

This function must be used with the [`DIFFERENTIAL_PRIVACY` clause][dp-syntax]
and can support these arguments:

+ `expression`: The input expression. This can be any numeric input type,
  such as `INT64`. `NULL` values are always ignored.
+ `contribution_bounds_per_group`: A named argument with a
  [contribution bound][dp-clamped-named]. Performs clamping for each group
  separately before performing intermediate grouping on the privacy unit column.

**Return type**

One of the following [supertypes][dp-supertype]:

+ `INT64`
+ `UINT64`
+ `DOUBLE`

**Examples**

The following differentially private query gets the sum of items requested.
Smaller aggregations might not be included. This query references a view called
[`professors`][dp-example-tables].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    SUM(quantity, contribution_bounds_per_group => (0,100)) quantity
FROM professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+-----------+
 | item     | quantity  |
 +----------+-----------+
 | pencil   | 143       |
 | pen      | 59        |
 +----------+-----------*/
```

```googlesql
-- Without noise, using the epsilon parameter.
-- (this un-noised version is for demonstration only)
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    SUM(quantity) quantity
FROM professors
GROUP BY item;

-- These results will not change when you run the query.
/*----------+----------+
 | item     | quantity |
 +----------+----------+
 | scissors | 8        |
 | pencil   | 144      |
 | pen      | 58       |
 +----------+----------*/
```

The following differentially private query gets the sum of items requested.
Smaller aggregations might not be included. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    SUM(quantity, contribution_bounds_per_group=>(0, 100)) quantity
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+-----------+
 | item     | quantity  |
 +----------+-----------+
 | pencil   | 143       |
 | pen      | 59        |
 +----------+-----------*/
```

```googlesql
-- Without noise, using the epsilon parameter.
-- (this un-noised version is for demonstration only)
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1)
    item,
    SUM(quantity) quantity
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will not change when you run the query.
/*----------+----------+
 | item     | quantity |
 +----------+----------+
 | scissors | 8        |
 | pencil   | 144      |
 | pen      | 58       |
 +----------+----------*/
```

Note: For more information about when and when not to use
noise, see [Use differential privacy][dp-noise].

[dp-example-tables]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_tables

[dp-noise]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#eliminate_noise

[dp-supertype]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md#supertypes

[dp-clamped-named]: #dp_clamped_named

[dp-syntax]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_clause

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
