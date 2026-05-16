---
name: ANON_PERCENTILE_CONT
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  Approximated by linear interpolation over the noised sorted clipped values.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#anon_percentile_cont-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/anon_percentile_cont.yaml
---

# ANON_PERCENTILE_CONT

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

## `ANON_PERCENTILE_CONT` (DEPRECATED) 
<a id="anon_percentile_cont"></a>

Warning: This function has been deprecated. Use
`PERCENTILE_CONT` (differential privacy) instead.

```googlesql
WITH ANONYMIZATION ...
  ANON_PERCENTILE_CONT(expression, percentile [CLAMPED BETWEEN lower_bound AND upper_bound])
```

**Description**

Takes an expression and computes a percentile for it. The final result is an
aggregation across privacy unit columns.

This function must be used with the `ANONYMIZATION` clause and
can support these arguments:

+ `expression`: The input expression. This can be most numeric input types,
  such as `INT64`. `NULL`s are always ignored.
+ `percentile`: The percentile to compute. The percentile must be a literal in
  the range [0, 1]
+ `CLAMPED BETWEEN` clause:
  Perform [clamping][dp-clamping] per privacy unit column.

`NUMERIC` and `BIGNUMERIC` arguments aren't allowed.
 If you need them, cast them as the
`DOUBLE` data type first.

**Return type**

`DOUBLE`

**Examples**

The following differentially private query gets the percentile of items
requested. Smaller aggregations might not be included. This query references a
view called [`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    ANON_PERCENTILE_CONT(quantity, 0.5 CLAMPED BETWEEN 0 AND 100) percentile_requested
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

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

[dp-clamping]: #dp_clamping

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
