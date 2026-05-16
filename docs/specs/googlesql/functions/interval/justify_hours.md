---
name: JUSTIFY_HOURS
dialect: googlesql
category: functions/interval
status: implemented
source_url: docs/third_party/googlesql-docs/interval_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/interval_functions.md#justify_hours
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/interval/justify_hours.yaml
---

# JUSTIFY_HOURS

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

## `JUSTIFY_HOURS`

```googlesql
JUSTIFY_HOURS(interval_expression)
```

**Description**

Normalizes the time part of the interval to the range from -23:59:59.999999 to
23:59:59.999999 by incrementing/decrementing the day part of the interval.

**Return Data Type**

`INTERVAL`

**Example**

```googlesql
SELECT
  JUSTIFY_HOURS(INTERVAL 23 HOUR) AS i1,
  JUSTIFY_HOURS(INTERVAL -24 HOUR) AS i2,
  JUSTIFY_HOURS(INTERVAL 47 HOUR) AS i3,
  JUSTIFY_HOURS(INTERVAL -12345 MINUTE) AS i4

/*--------------+--------------+--------------+-----------------+
 | i1           | i2           | i3           | i4              |
 +--------------+--------------+--------------+-----------------+
 | 0-0 0 23:0:0 | 0-0 -1 0:0:0 | 0-0 1 23:0:0 | 0-0 -8 -13:45:0 |
 +--------------+--------------+--------------+-----------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/interval_functions.md`.
