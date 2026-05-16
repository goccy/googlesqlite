---
name: DATETIME_DIFF
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#datetime_diff
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/datetime_diff.yaml
---

# DATETIME_DIFF

## Summary

Returns the number of unit boundaries between two `DATETIME` values
(`end_datetime` - `start_datetime`) at a specified time granularity.

## Signatures

- `DATETIME_DIFF(end_datetime, start_datetime, granularity)`

## Behavior

- Return type is `INT64`.
- Counts the number of `granularity` boundaries crossed between
  `start_datetime` and `end_datetime`, not the elapsed duration.
- Supported `granularity` values are `NANOSECOND`, `MICROSECOND`,
  `MILLISECOND`, `SECOND`, `MINUTE`, `HOUR`, `DAY`, `WEEK`,
  `WEEK(<WEEKDAY>)`, `ISOWEEK`, `MONTH`, `QUARTER`, `YEAR`, and
  `ISOYEAR`.
- `WEEK` boundaries begin on Sunday; `WEEK(<WEEKDAY>)` begins on the
  named weekday; `ISOWEEK` follows ISO 8601 and begins on Monday.
- `ISOYEAR` uses the ISO 8601 week-numbering year boundary (the
  Monday of the first week whose Thursday is in the corresponding
  Gregorian year).
- If `end_datetime` is earlier than `start_datetime`, the result is
  negative.
- The function follows the type of arguments passed in; for example,
  `DATETIME_DIFF(TIMESTAMP, TIMESTAMP, PART)` behaves like
  `TIMESTAMP_DIFF(TIMESTAMP, TIMESTAMP, PART)`.

## Examples

```googlesql
SELECT
  DATETIME_DIFF(DATETIME "2010-07-07 10:20:00",
    DATETIME "2008-12-25 15:30:00", DAY) AS difference;
-- expected: 559
```

```googlesql
SELECT
  DATETIME_DIFF(DATETIME '2017-10-15 00:00:00',
    DATETIME '2017-10-14 00:00:00', DAY) AS days_diff,
  DATETIME_DIFF(DATETIME '2017-10-15 00:00:00',
    DATETIME '2017-10-14 00:00:00', WEEK) AS weeks_diff;
-- expected: days_diff=1, weeks_diff=1
```

```googlesql
SELECT
  DATETIME_DIFF('2017-12-18', '2017-12-17', WEEK) AS week_diff,
  DATETIME_DIFF('2017-12-18', '2017-12-17', WEEK(MONDAY)) AS week_weekday_diff,
  DATETIME_DIFF('2017-12-18', '2017-12-17', ISOWEEK) AS isoweek_diff;
-- expected: week_diff=0, week_weekday_diff=1, isoweek_diff=1
```

## Edge cases

- Produces an error if the computation overflows, such as when the
  difference in nanoseconds between the two `DATETIME` values
  overflows `INT64`.
- Returns a negative value when `end_datetime` precedes
  `start_datetime`.
- Because the function counts boundary crossings rather than full
  units, two `DATETIME`s 24 hours apart that span a Saturday/Sunday
  midnight return `1` for `WEEK`.
- `YEAR` and `ISOYEAR` can disagree near year boundaries; for
  example, `2014-12-30` belongs to ISO year 2015, so the same pair
  of `DATETIME`s may yield a different `YEAR` vs `ISOYEAR`
  difference.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/datetime_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATETIME_DIFF`

```googlesql
DATETIME_DIFF(end_datetime, start_datetime, granularity)
```

**Description**

Gets the number of unit boundaries between two `DATETIME` values
(`end_datetime` - `start_datetime`) at a particular time granularity.

**Definitions**

+   `start_datetime`: The starting `DATETIME` value.
+   `end_datetime`: The ending `DATETIME` value.
+   `granularity`: The datetime part that represents the granularity. If
    you have passed in `DATETIME` values for the first arguments, `granularity`
    can be:

      
      + `NANOSECOND`
      + `MICROSECOND`
      + `MILLISECOND`
      + `SECOND`
      + `MINUTE`
      + `HOUR`
      + `DAY`
      + `WEEK`: This date part begins on Sunday.
      + `WEEK(<WEEKDAY>)`: This date part begins on `WEEKDAY`. Valid values for
        `WEEKDAY` are `SUNDAY`, `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`,
        `FRIDAY`, and `SATURDAY`.
      + `ISOWEEK`: Uses [ISO 8601 week][ISO-8601-week]
        boundaries. ISO weeks begin on Monday.
      + `MONTH`
      + `QUARTER`
      + `YEAR`
      + `ISOYEAR`: Uses the [ISO 8601][ISO-8601]
        week-numbering year boundary. The ISO year boundary is the Monday of the
        first week whose Thursday belongs to the corresponding Gregorian calendar
        year.

**Details**

If `end_datetime` is earlier than `start_datetime`, the output is negative.
Produces an error if the computation overflows, such as if the difference
in nanoseconds
between the two `DATETIME` values overflows.

Note: The behavior of the this function follows the type of arguments passed in.
For example, `DATETIME_DIFF(TIMESTAMP, TIMESTAMP, PART)`
behaves like `TIMESTAMP_DIFF(TIMESTAMP, TIMESTAMP, PART)`.

**Return Data Type**

`INT64`

**Example**

```googlesql
SELECT
  DATETIME "2010-07-07 10:20:00" as first_datetime,
  DATETIME "2008-12-25 15:30:00" as second_datetime,
  DATETIME_DIFF(DATETIME "2010-07-07 10:20:00",
    DATETIME "2008-12-25 15:30:00", DAY) as difference;

/*----------------------------+------------------------+------------------------+
 | first_datetime             | second_datetime        | difference             |
 +----------------------------+------------------------+------------------------+
 | 2010-07-07 10:20:00        | 2008-12-25 15:30:00    | 559                    |
 +----------------------------+------------------------+------------------------*/
```

```googlesql
SELECT
  DATETIME_DIFF(DATETIME '2017-10-15 00:00:00',
    DATETIME '2017-10-14 00:00:00', DAY) as days_diff,
  DATETIME_DIFF(DATETIME '2017-10-15 00:00:00',
    DATETIME '2017-10-14 00:00:00', WEEK) as weeks_diff;

/*-----------+------------+
 | days_diff | weeks_diff |
 +-----------+------------+
 | 1         | 1          |
 +-----------+------------*/
```

The example above shows the result of `DATETIME_DIFF` for two `DATETIME`s that
are 24 hours apart. `DATETIME_DIFF` with the part `WEEK` returns 1 because
`DATETIME_DIFF` counts the number of part boundaries in this range of
`DATETIME`s. Each `WEEK` begins on Sunday, so there is one part boundary between
Saturday, `2017-10-14 00:00:00` and Sunday, `2017-10-15 00:00:00`.

The following example shows the result of `DATETIME_DIFF` for two dates in
different years. `DATETIME_DIFF` with the date part `YEAR` returns 3 because it
counts the number of Gregorian calendar year boundaries between the two
`DATETIME`s. `DATETIME_DIFF` with the date part `ISOYEAR` returns 2 because the
second `DATETIME` belongs to the ISO year 2015. The first Thursday of the 2015
calendar year was 2015-01-01, so the ISO year 2015 begins on the preceding
Monday, 2014-12-29.

```googlesql
SELECT
  DATETIME_DIFF('2017-12-30 00:00:00',
    '2014-12-30 00:00:00', YEAR) AS year_diff,
  DATETIME_DIFF('2017-12-30 00:00:00',
    '2014-12-30 00:00:00', ISOYEAR) AS isoyear_diff;

/*-----------+--------------+
 | year_diff | isoyear_diff |
 +-----------+--------------+
 | 3         | 2            |
 +-----------+--------------*/
```

The following example shows the result of `DATETIME_DIFF` for two days in
succession. The first date falls on a Monday and the second date falls on a
Sunday. `DATETIME_DIFF` with the date part `WEEK` returns 0 because this time
part uses weeks that begin on Sunday. `DATETIME_DIFF` with the date part
`WEEK(MONDAY)` returns 1. `DATETIME_DIFF` with the date part
`ISOWEEK` also returns 1 because ISO weeks begin on Monday.

```googlesql
SELECT
  DATETIME_DIFF('2017-12-18', '2017-12-17', WEEK) AS week_diff,
  DATETIME_DIFF('2017-12-18', '2017-12-17', WEEK(MONDAY)) AS week_weekday_diff,
  DATETIME_DIFF('2017-12-18', '2017-12-17', ISOWEEK) AS isoweek_diff;

/*-----------+-------------------+--------------+
 | week_diff | week_weekday_diff | isoweek_diff |
 +-----------+-------------------+--------------+
 | 0         | 1                 | 1            |
 +-----------+-------------------+--------------*/
```

[ISO-8601]: https://en.wikipedia.org/wiki/ISO_8601

[ISO-8601-week]: https://en.wikipedia.org/wiki/ISO_week_date

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
