---
name: ANON_SUM
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  ANON_SUM alias of the differential-privacy SUM.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#anon_sum-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/anon_sum.yaml
---

# ANON_SUM

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

## `ANON_SUM` (DEPRECATED) 
<a id="anon_sum"></a>

Warning: This function has been deprecated. Use
`SUM` (differential privacy) instead.

```googlesql
WITH ANONYMIZATION ...
  ANON_SUM(expression [CLAMPED BETWEEN lower_bound AND upper_bound])
```

**Description**

Returns the sum of non-`NULL`, non-`NaN` values in the expression. The final
result is an aggregation across privacy unit columns.

This function must be used with the `ANONYMIZATION` clause and
can support these arguments:

+ `expression`: The input expression. This can be any numeric input type,
  such as `INT64`.
+ `CLAMPED BETWEEN` clause:
  Perform [clamping][dp-clamping] per privacy unit column.

**Return type**

One of the following [supertypes][dp-supertype]:

+ `INT64`
+ `UINT64`
+ `DOUBLE`

**Examples**

The following differentially private query gets the sum of items requested.
Smaller aggregations might not be included. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    ANON_SUM(quantity CLAMPED BETWEEN 0 AND 100) quantity
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
  WITH ANONYMIZATION
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1)
    item,
    ANON_SUM(quantity) quantity
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

Note: You can learn more about when and when not to use
noise [here][dp-noise].

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

[dp-noise]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#eliminate_noise

[dp-supertype]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md#supertypes

[dp-clamping]: #dp_clamping

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
