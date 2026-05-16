---
name: JUSTIFY_INTERVAL
dialect: googlesql
category: functions/interval
status: implemented
source_url: docs/third_party/googlesql-docs/interval_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/interval_functions.md#justify_interval
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/interval/justify_interval.yaml
---

# JUSTIFY_INTERVAL

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

Verbatim copy from `docs/third_party/googlesql-docs/interval_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `JUSTIFY_INTERVAL`

```googlesql
JUSTIFY_INTERVAL(interval_expression)
```

**Description**

Normalizes the days and time parts of the interval.

**Return Data Type**

`INTERVAL`

**Example**

```googlesql
SELECT JUSTIFY_INTERVAL(INTERVAL '29 49:00:00' DAY TO SECOND) AS i

/*-------------+
 | i           |
 +-------------+
 | 0-1 1 1:0:0 |
 +-------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/interval_functions.md`.
