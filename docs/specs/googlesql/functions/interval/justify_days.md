---
name: JUSTIFY_DAYS
dialect: googlesql
category: functions/interval
status: implemented
source_url: docs/third_party/googlesql-docs/interval_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/interval_functions.md#justify_days
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/interval/justify_days.yaml
---

# JUSTIFY_DAYS

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

## `JUSTIFY_DAYS`

```googlesql
JUSTIFY_DAYS(interval_expression)
```

**Description**

Normalizes the day part of the interval to the range from -29 to 29 by
incrementing/decrementing the month or year part of the interval.

**Return Data Type**

`INTERVAL`

**Example**

```googlesql
SELECT
  JUSTIFY_DAYS(INTERVAL 29 DAY) AS i1,
  JUSTIFY_DAYS(INTERVAL -30 DAY) AS i2,
  JUSTIFY_DAYS(INTERVAL 31 DAY) AS i3,
  JUSTIFY_DAYS(INTERVAL -65 DAY) AS i4,
  JUSTIFY_DAYS(INTERVAL 370 DAY) AS i5

/*--------------+--------------+-------------+---------------+--------------+
 | i1           | i2           | i3          | i4            | i5           |
 +--------------+--------------+-------------+---------------+--------------+
 | 0-0 29 0:0:0 | -0-1 0 0:0:0 | 0-1 1 0:0:0 | -0-2 -5 0:0:0 | 1-0 10 0:0:0 |
 +--------------+--------------+-------------+---------------+--------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/interval_functions.md`.
