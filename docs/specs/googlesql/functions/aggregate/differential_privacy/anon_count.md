---
name: ANON_COUNT
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  ANON_COUNT runs through the analyzer's Anonymization rewriter,
  which lowers `SELECT WITH ANONYMIZATION` into a chain of partial
  aggregates that route to the `googlesqlite_anon_count_star` /
  `googlesqlite_anon_count` runtimes via AnonymizedAggregateScan +
  SampleScan + AggregateScan. SampleScan now passes through to its
  InputScan, AnonymizedAggregateScanNode emits the same SELECT shape
  as DifferentialPrivacyAggregateScan, and the AggregateFunctionCall
  formatter appends (epsilon, delta) args to every
  `googlesqlite_anon_*` call so the runtime can size the Laplace
  noise from the scan's privacy budget. Source tables surface their
  `OPTIONS(anonymization_userid_column='id')` via
  SimpleTable.SetAnonymizationInfo2 (catalog plumbing in
  applyTableOptions).
  
  Testdata only asserts the high-epsilon (`epsilon=1e20`) Examples
  because the docs' noisy Examples (`epsilon=10`) carry the
  explicit caveat "These results will change each time you run the
  query" — those specific values are a single noise realization and
  cannot be a regression assertion. The deterministic Examples
  exercise the full analyzer / formatter / runtime path including
  the CLAMPED BETWEEN bounds.
source_url: docs/third_party/googlesql-docs/aggregate-dp-functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#anon_count-deprecated
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/anon_count.yaml
---

# ANON_COUNT

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

## `ANON_COUNT` (DEPRECATED) 
<a id="anon_count"></a>

Warning: This function has been deprecated. Use
`COUNT` (differential privacy) instead.

+ [Signature 1](#anon_count_signature1)
+ [Signature 2](#anon_count_signature2)

#### Signature 1 
<a id="anon_count_signature1"></a>

```googlesql
WITH ANONYMIZATION ...
  ANON_COUNT(*)
```

**Description**

Returns the number of rows in the
[differentially private][dp-from-clause] `FROM` clause. The final result
is an aggregation across privacy unit columns.
[Input values are clamped implicitly][dp-clamp-implicit]. Clamping is
performed per privacy unit column.

This function must be used with the `ANONYMIZATION` clause.

**Return type**

`INT64`

**Examples**

The following differentially private query counts the number of requests for
each item. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    ANON_COUNT(*) times_requested
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+-----------------+
 | item     | times_requested |
 +----------+-----------------+
 | pencil   | 5               |
 | pen      | 2               |
 +----------+-----------------*/
```

```googlesql
-- Without noise, using the epsilon parameter.
-- (this un-noised version is for demonstration only)
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1)
    item,
    ANON_COUNT(*) times_requested
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will not change when you run the query.
/*----------+-----------------+
 | item     | times_requested |
 +----------+-----------------+
 | scissors | 1               |
 | pencil   | 4               |
 | pen      | 3               |
 +----------+-----------------*/
```

Note: You can learn more about when and when not to use
noise [here][dp-noise].

#### Signature 2 
<a id="anon_count_signature2"></a>

```googlesql
WITH ANONYMIZATION ...
  ANON_COUNT(expression [CLAMPED BETWEEN lower_bound AND upper_bound])
```

**Description**

Returns the number of non-`NULL` expression values. The final result is an
aggregation across privacy unit columns.

This function must be used with the `ANONYMIZATION` clause and
can support these arguments:

+ `expression`: The input expression. This can be any numeric input type,
  such as `INT64`.
+ `CLAMPED BETWEEN` clause:
  Perform [clamping][dp-clamping] per privacy unit column.

**Return type**

`INT64`

**Examples**

The following differentially private query counts the number of requests made
for each type of item. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    ANON_COUNT(item CLAMPED BETWEEN 0 AND 100) times_requested
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will change each time you run the query.
-- Smaller aggregations might be removed.
/*----------+-----------------+
 | item     | times_requested |
 +----------+-----------------+
 | pencil   | 5               |
 | pen      | 2               |
 +----------+-----------------*/
```

```googlesql
--Without noise (this un-noised version is for demonstration only)
SELECT
  WITH ANONYMIZATION
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1)
    item,
    ANON_COUNT(item CLAMPED BETWEEN 0 AND 100) times_requested
FROM {{USERNAME}}.view_on_professors
GROUP BY item;

-- These results will not change when you run the query.
/*----------+-----------------+
 | item     | times_requested |
 +----------+-----------------+
 | scissors | 1               |
 | pencil   | 4               |
 | pen      | 3               |
 +----------+-----------------*/
```

Note: You can learn more about when and when not to use
noise [here][dp-noise].

[dp-clamp-implicit]: #dp_clamped_named_implicit

[dp-from-clause]: https://github.com/google/googlesql/blob/master/docs/differential-privacy.md#dp_from_rules

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

[dp-noise]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#eliminate_noise

[dp-clamping]: #dp_clamping

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
