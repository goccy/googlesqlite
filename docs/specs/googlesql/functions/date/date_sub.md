---
name: DATE_SUB
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#date_sub
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/date_sub.yaml
---

# DATE_SUB

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

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATE_SUB`

```googlesql
DATE_SUB(date_expression, INTERVAL int64_expression date_part)
```

**Description**

Subtracts a specified time interval from a DATE.

`DATE_SUB` supports the following `date_part` values:

+  `DAY`
+  `WEEK`. Equivalent to 7 `DAY`s.
+  `MONTH`
+  `QUARTER`
+  `YEAR`

Special handling is required for MONTH, QUARTER, and YEAR parts when
the date is at (or near) the last day of the month. If the resulting
month has fewer days than the original date's day, then the resulting
date is the last date of that month.

**Return Data Type**

DATE

**Example**

```googlesql
SELECT DATE_SUB(DATE '2008-12-25', INTERVAL 5 DAY) AS five_days_ago;

/*---------------+
 | five_days_ago |
 +---------------+
 | 2008-12-20    |
 +---------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
