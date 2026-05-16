---
name: HLL_COUNT.MERGE
dialect: googlesql
category: functions/hll
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/hll_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hll_functions.md#hll_countmerge
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hll/hll_count_merge.yaml
---

# HLL_COUNT.MERGE

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

## `HLL_COUNT.MERGE`

```
HLL_COUNT.MERGE(sketch)
```

**Description**

An aggregate function that returns the cardinality of several
[HLL++][hll-link-to-research-whitepaper] sketches by computing their union.

Each `sketch` must be initialized on the same type. Attempts to merge sketches
for different types results in an error. For example, you can't merge a sketch
initialized from `INT64` data with one initialized from `STRING` data.

If the merged sketches were initialized with different precisions, the precision
will be downgraded to the lowest precision involved in the merge.

This function ignores `NULL` values when merging sketches. If the merge happens
over zero rows or only over `NULL` values, the function returns `0`.

**Supported input types**

`BYTES`

**Return type**

`INT64`

**Example**

 The following query counts the number of distinct users across all countries
 who have at least one invoice.

```googlesql
SELECT HLL_COUNT.MERGE(hll_sketch) AS distinct_customers_with_open_invoice
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

/*--------------------------------------+
 | distinct_customers_with_open_invoice |
 +--------------------------------------+
 |                                    3 |
 +--------------------------------------*/
```

[hll-link-to-research-whitepaper]: https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/40671.pdf

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hll_functions.md`.
