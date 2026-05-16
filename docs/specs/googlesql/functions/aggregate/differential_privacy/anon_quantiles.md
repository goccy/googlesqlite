---
name: ANON_QUANTILES
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  Approximated by post-noise sort of the clipped values.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#anon_quantiles-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/anon_quantiles.yaml
---

# ANON_QUANTILES

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

## `ANON_QUANTILES` (DEPRECATED) 
<a id="anon_quantiles"></a>

Warning: This function has been deprecated. Use
`QUANTILES` (differential privacy) instead.

```googlesql
WITH ANONYMIZATION ...
  ANON_QUANTILES(expression, number CLAMPED BETWEEN lower_bound AND upper_bound)
```

**Description**

Returns an array of differentially private quantile boundaries for values in
`expression`. The first element in the return value is the
minimum quantile boundary and the last element is the maximum quantile boundary.
The returned results are aggregations across privacy unit columns.

This function must be used with the `ANONYMIZATION` clause and
can support these arguments:

+ `expression`: The input expression. This can be most numeric input types,
  such as `INT64`. `NULL`s are always ignored.
+ `number`: The number of quantiles to create. This must be an `INT64`.
+ `CLAMPED BETWEEN` clause:
  Perform [clamping][dp-clamping] per privacy unit column.

`NUMERIC` and `BIGNUMERIC` arguments aren't allowed.
 If you need them, cast them as the
`DOUBLE` data type first.

**Return type**

`ARRAY`\<`DOUBLE`>

**Examples**

The following differentially private query gets the five quantile boundaries of
the four quartiles of the number of items requested. Smaller aggregations
might not be included. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    ANON_QUANTILES(quantity, 4 CLAMPED BETWEEN 0 AND 100) quantiles_requested
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+----------------------------------------------------------------------+
 | item     | quantiles_requested                                                  |
 +----------+----------------------------------------------------------------------+
 | pen      | [6.409375,20.647684733072918,41.40625,67.30848524305556,99.80078125] |
 | pencil   | [6.849259,44.010416666666664,62.64204,65.83806818181819,98.59375]    |
 +----------+----------------------------------------------------------------------*/
```

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

[dp-clamping]: #dp_clamping

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
