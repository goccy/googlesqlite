---
name: DATE_ADD
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#date_add
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/date_add.yaml
---

# DATE_ADD

## Summary

Adds a specified time interval to a DATE and returns the resulting DATE.

## Signatures

- ```googlesql
  DATE_ADD(date_expression, INTERVAL int64_expression date_part)
  ```

## Behavior

- Returns a `DATE`.
- Adds `int64_expression` units of `date_part` to `date_expression`.
- Supported `date_part` values are `DAY`, `WEEK`, `MONTH`, `QUARTER`, and `YEAR`.
- `WEEK` is equivalent to 7 `DAY`s.
- For `MONTH`, `QUARTER`, and `YEAR`, when the resulting month has fewer days than the original date's day, the result is clamped to the last day of the resulting month.

## Examples

```googlesql
SELECT DATE_ADD(DATE '2008-12-25', INTERVAL 5 DAY) AS five_days_later;
-- expected: 2008-12-30
```

## Edge cases

- When adding `MONTH`, `QUARTER`, or `YEAR` to a date at or near the end of a month, the day is clamped to the last day of the resulting month if that month has fewer days (e.g. Jan 31 + 1 MONTH yields the last day of February).
- `date_part` values other than `DAY`, `WEEK`, `MONTH`, `QUARTER`, and `YEAR` are not supported.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATE_ADD`

```googlesql
DATE_ADD(date_expression, INTERVAL int64_expression date_part)
```

**Description**

Adds a specified time interval to a DATE.

`DATE_ADD` supports the following `date_part` values:

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
SELECT DATE_ADD(DATE '2008-12-25', INTERVAL 5 DAY) AS five_days_later;

/*--------------------+
 | five_days_later    |
 +--------------------+
 | 2008-12-30         |
 +--------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
