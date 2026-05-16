---
name: LAST_DAY
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#last_day
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/last_day.yaml
---

# LAST_DAY

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

Verbatim copy from `docs/third_party/googlesql-docs/datetime_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `LAST_DAY`

```googlesql
LAST_DAY(datetime_expression[, date_part])
```

**Description**

Returns the last day from a datetime expression that contains the date.
This is commonly used to return the last day of the month.

You can optionally specify the date part for which the last day is returned.
If this parameter isn't used, the default value is `MONTH`.
`LAST_DAY` supports the following values for `date_part`:

+  `YEAR`
+  `QUARTER`
+  `MONTH`
+  `WEEK`. Equivalent to 7 `DAY`s.
+  `WEEK(<WEEKDAY>)`. `<WEEKDAY>` represents the starting day of the week.
   Valid values are `SUNDAY`, `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`,
   `FRIDAY`, and `SATURDAY`.
+  `ISOWEEK`. Uses [ISO 8601][ISO-8601-week] week boundaries. ISO weeks begin
   on Monday.
+  `ISOYEAR`. Uses the [ISO 8601][ISO-8601] week-numbering year boundary.
   The ISO year boundary is the Monday of the first week whose Thursday belongs
   to the corresponding Gregorian calendar year.

**Return Data Type**

`DATE`

**Example**

These both return the last day of the month:

```googlesql
SELECT LAST_DAY(DATETIME '2008-11-25', MONTH) AS last_day

/*------------+
 | last_day   |
 +------------+
 | 2008-11-30 |
 +------------*/
```

```googlesql
SELECT LAST_DAY(DATETIME '2008-11-25') AS last_day

/*------------+
 | last_day   |
 +------------+
 | 2008-11-30 |
 +------------*/
```

This returns the last day of the year:

```googlesql
SELECT LAST_DAY(DATETIME '2008-11-25 15:30:00', YEAR) AS last_day

/*------------+
 | last_day   |
 +------------+
 | 2008-12-31 |
 +------------*/
```

This returns the last day of the week for a week that starts on a Sunday:

```googlesql
SELECT LAST_DAY(DATETIME '2008-11-10 15:30:00', WEEK(SUNDAY)) AS last_day

/*------------+
 | last_day   |
 +------------+
 | 2008-11-15 |
 +------------*/
```

This returns the last day of the week for a week that starts on a Monday:

```googlesql
SELECT LAST_DAY(DATETIME '2008-11-10 15:30:00', WEEK(MONDAY)) AS last_day

/*------------+
 | last_day   |
 +------------+
 | 2008-11-16 |
 +------------*/
```

[ISO-8601]: https://en.wikipedia.org/wiki/ISO_8601

[ISO-8601-week]: https://en.wikipedia.org/wiki/ISO_week_date

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
