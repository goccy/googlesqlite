---
name: HLL_COUNT.EXTRACT
dialect: googlesql
category: functions/hll
status: implemented
source_url: docs/third_party/googlesql-docs/hll_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hll_functions.md#hll_countextract
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hll/hll_count_extract.yaml
---

# HLL_COUNT.EXTRACT

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

## `HLL_COUNT.EXTRACT`

```
HLL_COUNT.EXTRACT(sketch)
```

**Description**

A scalar function that extracts a cardinality estimate of a single
[HLL++][hll-link-to-research-whitepaper] sketch.

If `sketch` is `NULL`, this function returns a cardinality estimate of `0`.

**Supported input types**

`BYTES`

**Return type**

`INT64`

**Example**

The following query returns the number of distinct users for each country who
have at least one invoice.

```googlesql
SELECT
  country,
  HLL_COUNT.EXTRACT(HLL_sketch) AS distinct_customers_with_open_invoice
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

/*---------+--------------------------------------+
 | country | distinct_customers_with_open_invoice |
 +---------+--------------------------------------+
 | UA      |                                    2 |
 | BR      |                                    1 |
 | CZ      |                                    1 |
 +---------+--------------------------------------*/
```

[hll-link-to-research-whitepaper]: https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/40671.pdf

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hll_functions.md`.
