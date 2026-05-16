---
name: STDDEV
dialect: googlesql
category: functions/aggregate/statistical
status: implemented
source_url: docs/third_party/googlesql-docs/statistical_aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/statistical_aggregate_functions.md#stddev
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/statistical/stddev.yaml
---

# STDDEV

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

Verbatim copy from `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `STDDEV`

```googlesql
STDDEV(
  [ DISTINCT ]
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
)
[ OVER over_clause ]

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  [ ORDER BY expression [ { ASC | DESC }  ] [, ...] ]
  [ window_frame_clause ]

```

**Description**

An alias of [STDDEV_SAMP][stat-agg-link-to-stddev-samp].

[stat-agg-link-to-stddev-samp]: #stddev_samp

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/statistical_aggregate_functions.md`.
