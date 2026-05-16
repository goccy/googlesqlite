---
name: ANON_STDDEV_POP
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  ANON_STDDEV_POP alias of the differential-privacy STDDEV_POP.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#anon_stddev_pop-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/anon_stddev_pop.yaml
---

# ANON_STDDEV_POP

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

## `ANON_STDDEV_POP` (DEPRECATED) 
<a id="anon_stddev_pop"></a>

Warning: This function has been deprecated. Use
`STDDEV_POP` (differential privacy) instead.

```googlesql
WITH ANONYMIZATION ...
  ANON_STDDEV_POP(expression [CLAMPED BETWEEN lower_bound AND upper_bound])
```

**Description**

Takes an expression and computes the population (biased) standard deviation of
the values in the expression. The final result is an aggregation across
privacy unit columns between `0` and `+Inf`.

This function must be used with the `ANONYMIZATION` clause and
can support these arguments:

+ `expression`: The input expression. This can be most numeric input types,
  such as `INT64`. `NULL`s are always ignored.
+ `CLAMPED BETWEEN` clause:
  Perform [clamping][dp-clamping] per individual entity values.

`NUMERIC` and `BIGNUMERIC` arguments aren't allowed.
 If you need them, cast them as the
`DOUBLE` data type first.

**Return type**

`DOUBLE`

**Examples**

The following differentially private query gets the
population (biased) standard deviation of items requested. Smaller aggregations
might not be included. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    ANON_STDDEV_POP(quantity CLAMPED BETWEEN 0 AND 100) pop_standard_deviation
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+------------------------+
 | item     | pop_standard_deviation |
 +----------+------------------------+
 | pencil   | 25.350871122442054     |
 | scissors | 50                     |
 | pen      | 2                      |
 +----------+------------------------*/
```

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

[dp-clamping]: #dp_clamping

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
