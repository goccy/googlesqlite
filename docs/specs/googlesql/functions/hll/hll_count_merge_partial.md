---
name: HLL_COUNT.MERGE_PARTIAL
dialect: googlesql
category: functions/hll
status: partial
notes: |
  The aggregate evaluates end-to-end against the upstream Example
  query: HLL_COUNT.INIT seeds three sketches (one per country) and
  HLL_COUNT.MERGE_PARTIAL folds them into a single sketch that
  HLL_COUNT.EXTRACT applied subsequently returns the right
  cardinality. The blocker is wire-format parity: the upstream
  Example's expected output is the BigQuery-internal `HllSketchProto`
  serialisation
  `"\010p\020\006\030\002 \013\202\007\020\020\003\030\017 \0242\010\320\2408\352}\244\223\002"`,
  which encodes (num_values, encoding_version, register_width,
  precision) varints plus a length-delimited sparse register table
  carrying delta-coded register values. Our runtime uses
  `DataDog/go-hll` for the merge / register storage and emits its
  own byte format. Closing this case requires writing a BigQuery
  `HllSketchProto` encoder + register-merge implementation that
  matches the upstream wire layout byte-for-byte; there is no
  public Go implementation of that format. That work is a sizeable
  isolated component, tracked separately from this sweep. The
  spec stays partial until the format implementation lands.
source_url: docs/third_party/googlesql-docs/hll_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hll_functions.md#hll_countmerge_partial
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hll/hll_count_merge_partial.yaml
---

# HLL_COUNT.MERGE_PARTIAL

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

Verbatim copy from `docs/third_party/googlesql-docs/hll_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `HLL_COUNT.MERGE_PARTIAL`

```
HLL_COUNT.MERGE_PARTIAL(sketch)
```

**Description**

An aggregate function that takes one or more
[HLL++][hll-link-to-research-whitepaper] `sketch`
inputs and merges them into a new sketch.

Each `sketch` must be initialized on the same type. Attempts to merge sketches
for different types results in an error. For example, you can't merge a sketch
initialized from `INT64` data with one initialized from `STRING` data.

If the merged sketches were initialized with different precisions, the precision
will be downgraded to the lowest precision involved in the merge. For example,
if `MERGE_PARTIAL` encounters sketches of precision 14 and 15, the returned new
sketch will have precision 14.

This function returns `NULL` if there is no input or all inputs are `NULL`.

**Supported input types**

`BYTES`

**Return type**

`BYTES`

**Example**

The following query returns an HLL++ sketch that counts the number of distinct
users who have at least one invoice across all countries.

```googlesql
SELECT HLL_COUNT.MERGE_PARTIAL(HLL_sketch) AS distinct_customers_with_open_invoice
FROM
  (
    SELECT
      country,
      HLL_COUNT.INIT(customer_id) AS hll_sketch
    FROM
      UNNEST(
        ARRAY<STRUCT<country STRING, customer_id STRING, invoice_id STRING>>[
          ('UA', 'customer_id_1', 'invoice_id_11'),
          ('BR', 'customer_id_3', 'invoice_id_31'),
          ('CZ', 'customer_id_2', 'invoice_id_22'),
          ('CZ', 'customer_id_2', 'invoice_id_23'),
          ('BR', 'customer_id_3', 'invoice_id_31'),
          ('UA', 'customer_id_2', 'invoice_id_24')])
    GROUP BY country
  );

/*----------------------------------------------------------------------------------------------+
 | distinct_customers_with_open_invoice                                                         |
 +----------------------------------------------------------------------------------------------+
 | "\010p\020\006\030\002 \013\202\007\020\020\003\030\017 \0242\010\320\2408\352}\244\223\002" |
 +----------------------------------------------------------------------------------------------*/
```

[hll-link-to-research-whitepaper]: https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/40671.pdf

[hll-sketches]: https://en.wikipedia.org/wiki/HyperLogLog

[cardinality]: https://en.wikipedia.org/wiki/Cardinality

[count-distinct]: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#count

[approx-count-distinct]: https://github.com/google/googlesql/blob/master/docs/approximate_aggregate_functions.md#approx_count_distinct

[approx-functions-reference]: https://github.com/google/googlesql/blob/master/docs/approximate_aggregate_functions.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hll_functions.md`.
