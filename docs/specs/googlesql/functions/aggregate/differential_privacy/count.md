---
name: COUNT
dialect: googlesql
category: functions/aggregate/differential_privacy
status: implemented
notes: |
  Counts per-row contributions and adds Laplace noise with scale = max(|lo|, |hi|) / epsilon. Bounds default to (0, 1) for the per-row case.
  
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
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate-dp-functions.md#count-differential_privacy
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/differential_privacy/count.yaml
---

# COUNT

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

## `COUNT` (`DIFFERENTIAL_PRIVACY`) 
<a id="dp_count"></a>

+ [Signature 1](#dp_count_signature1): Returns the number of rows in a
  differentially private `FROM` clause.
+ [Signature 2](#dp_count_signature2): Returns the number of non-`NULL`
  values in an expression.

#### Signature 1 
<a id="dp_count_signature1"></a>

```googlesql
WITH DIFFERENTIAL_PRIVACY ...
  COUNT(
    *,
    [ contribution_bounds_per_group => (lower_bound, upper_bound) ]
  )
```

**Description**

Returns the number of rows in the
[differentially private][dp-from-clause] `FROM` clause. The final result
is an aggregation across a privacy unit column.

This function must be used with the [`DIFFERENTIAL_PRIVACY` clause][dp-syntax]
and can support the following arguments:

+ `contribution_bounds_per_group`: A named argument with a
  [contribution bound][dp-clamped-named].
  Performs clamping for each group separately before performing intermediate
  grouping on the privacy unit column.

**Return type**

`INT64`

**Examples**

The following differentially private query counts the number of requests for
each item. This query references a table called
[`professors`][dp-example-tables].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    COUNT(*, contribution_bounds_per_group=>(0, 100)) times_requested
FROM professors
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
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    COUNT(*, contribution_bounds_per_group=>(0, 100)) times_requested
FROM professors
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

The following differentially private query counts the number of requests for
each item. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    COUNT(*, contribution_bounds_per_group=>(0, 100)) times_requested
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
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1)
    item,
    COUNT(*, contribution_bounds_per_group=>(0, 100)) times_requested
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

Note: For more information about when and when not to use
noise, see [Remove noise][dp-noise].

#### Signature 2 
<a id="dp_count_signature2"></a>

```googlesql
WITH DIFFERENTIAL_PRIVACY ...
  COUNT(
    expression,
    [contribution_bounds_per_group => (lower_bound, upper_bound)]
  )
```

**Description**

Returns the number of non-`NULL` expression values. The final result is an
aggregation across a privacy unit column.

This function must be used with the [`DIFFERENTIAL_PRIVACY` clause][dp-syntax]
and can support these arguments:

+ `expression`: The input expression. This expression can be any
  numeric input type, such as `INT64`.
+ `contribution_bounds_per_group`: A named argument with a
  [contribution bound][dp-clamped-named].
  Performs clamping per each group separately before performing intermediate
  grouping on the privacy unit column.

**Return type**

`INT64`

**Examples**

The following differentially private query counts the number of requests made
for each type of item. This query references a table called
[`professors`][dp-example-tables].

```googlesql
-- With noise, using the epsilon parameter.
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    COUNT(item, contribution_bounds_per_group => (0,100)) times_requested
FROM professors
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
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1, privacy_unit_column=id)
    item,
    COUNT(item, contribution_bounds_per_group => (0,100)) times_requested
FROM professors
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

The following differentially private query counts the number of requests made
for each type of item. This query references a view called
[`view_on_professors`][dp-example-views].

```googlesql
-- With noise
SELECT
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=10, delta=.01, max_groups_contributed=1)
    item,
    COUNT(item, contribution_bounds_per_group=>(0, 100)) times_requested
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
  WITH DIFFERENTIAL_PRIVACY
    OPTIONS(epsilon=1e20, delta=.01, max_groups_contributed=1)
    item,
    COUNT(item, contribution_bounds_per_group=>(0, 100)) times_requested
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

Note: For more information about when and when not to use
noise, see [Remove noise][dp-noise].

[dp-clamp-implicit]: #dp_implicit_clamping

[dp-from-clause]: https://github.com/google/googlesql/blob/master/docs/differential-privacy.md#dp_from_rules

[dp-example-tables]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_tables

[dp-noise]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#eliminate_noise

[dp-syntax]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_clause

[dp-clamped-named]: #dp_clamped_named

[dp-example-views]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#dp_example_views

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate-dp-functions.md`.
