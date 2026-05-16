---
name: DATETIME_ADD
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#datetime_add
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/datetime_add.yaml
---

# DATETIME_ADD

## Summary

Adds a signed integer number of `part` units to a `DATETIME` value and returns the resulting `DATETIME`.

## Signatures

- `DATETIME_ADD(datetime_expression, INTERVAL int64_expression part)`

## Behavior

- Returns a `DATETIME`.
- Adds `int64_expression` units of `part` to `datetime_expression`.
- Supported `part` values: `NANOSECOND`, `MICROSECOND`, `MILLISECOND`, `SECOND`, `MINUTE`, `HOUR`, `DAY`, `WEEK`, `MONTH`, `QUARTER`, `YEAR`.
- `WEEK` is equivalent to 7 `DAY`s.
- For `MONTH`, `QUARTER`, and `YEAR` parts, when the source date is at (or near) the last day of the month and the resulting month has fewer days than the original day-of-month, the result day is clamped to the last day of the new month.

## Examples

```googlesql
SELECT
  DATETIME "2008-12-25 15:30:00" as original_date,
  DATETIME_ADD(DATETIME "2008-12-25 15:30:00", INTERVAL 10 MINUTE) as later;
-- expected later = 2008-12-25 15:40:00
```

## Edge cases

- For `MONTH`, `QUARTER`, or `YEAR` adjustments where the target month has fewer days than the source day-of-month, the day is truncated to the last day of the target month (e.g. adding 1 `MONTH` to `2008-01-31` yields `2008-02-29`).
- A negative `int64_expression` performs subtraction by the equivalent number of units.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/datetime_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATETIME_ADD`

```googlesql
DATETIME_ADD(datetime_expression, INTERVAL int64_expression part)
```

**Description**

Adds `int64_expression` units of `part` to the `DATETIME` object.

`DATETIME_ADD` supports the following values for `part`:

+ `NANOSECOND`
+ `MICROSECOND`
+ `MILLISECOND`
+ `SECOND`
+ `MINUTE`
+ `HOUR`
+ `DAY`
+ `WEEK`. Equivalent to 7 `DAY`s.
+ `MONTH`
+ `QUARTER`
+ `YEAR`

Special handling is required for MONTH, QUARTER, and YEAR parts when the
date is at (or near) the last day of the month. If the resulting month has fewer
days than the original DATETIME's day, then the result day is the last day of
the new month.

**Return Data Type**

`DATETIME`

**Example**

```googlesql
SELECT
  DATETIME "2008-12-25 15:30:00" as original_date,
  DATETIME_ADD(DATETIME "2008-12-25 15:30:00", INTERVAL 10 MINUTE) as later;

/*-----------------------------+------------------------+
 | original_date               | later                  |
 +-----------------------------+------------------------+
 | 2008-12-25 15:30:00         | 2008-12-25 15:40:00    |
 +-----------------------------+------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
