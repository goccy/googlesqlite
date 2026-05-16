---
name: HLL_COUNT.INIT
dialect: googlesql
category: functions/hll
status: implemented
source_url: docs/third_party/googlesql-docs/hll_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hll_functions.md#hll_countinit
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hll/hll_count_init.yaml
---

# HLL_COUNT.INIT

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

## `HLL_COUNT.INIT`

```
HLL_COUNT.INIT(input [, precision])
```

**Description**

An aggregate function that takes one or more `input` values and aggregates them
into a [HLL++][hll-link-to-research-whitepaper] sketch. Each sketch
is represented using the `BYTES` data type. You can then merge sketches using
`HLL_COUNT.MERGE` or `HLL_COUNT.MERGE_PARTIAL`. If no merging is needed,
you can extract the final count of distinct values from the sketch using
`HLL_COUNT.EXTRACT`.

This function supports an optional parameter, `precision`. This parameter
defines the accuracy of the estimate at the cost of additional memory required
to process the sketches or store them on disk. The range for this value is
`10` to `24`. The default value is `15`. For more information about precision,
see [Precision for sketches][precision_hll].

If the input is `NULL`, this function returns `NULL`.

For more information, see [HyperLogLog in Practice: Algorithmic Engineering of
a State of The Art Cardinality Estimation Algorithm][hll-link-to-research-whitepaper].

**Supported input types**

+ `INT64`
+ `UINT64`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `STRING`
+ `BYTES`

**Return type**

`BYTES`

**Example**

The following query creates HLL++ sketches that count the number of distinct
users with at least one invoice per country.

```googlesql
SELECT
  country,
  HLL_COUNT.INIT(customer_id, 10)
    AS hll_sketch
FROM
  UNNEST(
    ARRAY<STRUCT<country STRING, customer_id STRING, invoice_id STRING>>[
      ('UA', 'customer_id_1', 'invoice_id_11'),
      ('CZ', 'customer_id_2', 'invoice_id_22'),
      ('CZ', 'customer_id_2', 'invoice_id_23'),
      ('BR', 'customer_id_3', 'invoice_id_31'),
      ('UA', 'customer_id_2', 'invoice_id_24')])
GROUP BY country;

/*---------+------------------------------------------------------------------------------------+
 | country | hll_sketch                                                                         |
 +---------+------------------------------------------------------------------------------------+
 | UA      | "\010p\020\002\030\002 \013\202\007\r\020\002\030\n \0172\005\371\344\001\315\010" |
 | CZ      | "\010p\020\002\030\002 \013\202\007\013\020\001\030\n \0172\003\371\344\001"       |
 | BR      | "\010p\020\001\030\002 \013\202\007\013\020\001\030\n \0172\003\202\341\001"       |
 +---------+------------------------------------------------------------------------------------*/
```

[hll-link-to-research-whitepaper]: https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/40671.pdf

[precision_hll]: https://github.com/google/googlesql/blob/master/docs/sketches.md#precision_hll

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hll_functions.md`.
