---
name: DATETIME_TRUNC
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#datetime_trunc
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/datetime_trunc.yaml
---

# DATETIME_TRUNC

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

## `DATETIME_TRUNC`

```googlesql
DATETIME_TRUNC(datetime_value, datetime_granularity)
```

```googlesql
DATETIME_TRUNC(timestamp_value, timestamp_granularity[, time_zone])
```

**Description**

Truncates a `DATETIME` or `TIMESTAMP` value at a particular granularity.

**Definitions**

+ `datetime_value`: A `DATETIME` value to truncate.
+ `datetime_granularity`: The truncation granularity for a `DATETIME` value.
  [Date granularities][datetime-trunc-granularity-date] and
  [time granularities][datetime-trunc-granularity-time] can be used.
+ `timestamp_value`: A `TIMESTAMP` value to truncate.
+ `timestamp_granularity`: The truncation granularity for a `TIMESTAMP` value.
  [Date granularities][datetime-trunc-granularity-date] and
  [time granularities][datetime-trunc-granularity-time] can be used.
+ `time_zone`: A time zone to use with the `TIMESTAMP` value.
  [Time zone parts][datetime-time-zone-parts] can be used.
  Use this argument if you want to use a time zone other than
  the default time zone, which is implementation defined, as part of the
  truncate operation.

      Note: When truncating a timestamp to `MINUTE`
    or `HOUR` parts, this function determines the civil time of the
    timestamp in the specified (or default) time zone
    and subtracts the minutes and seconds (when truncating to `HOUR`) or the
    seconds (when truncating to `MINUTE`) from that timestamp.
    While this provides intuitive results in most cases, the result is
    non-intuitive near daylight savings transitions that aren't hour-aligned.

<a id="datetime_trunc_granularity_date"></a>

**Date granularity definitions**

  + `DAY`: The day in the Gregorian calendar year that contains the
    value to truncate.

  + `WEEK`: The first day in the week that contains the
    value to truncate. Weeks begin on Sundays. `WEEK` is equivalent to
    `WEEK(SUNDAY)`.

  + `WEEK(WEEKDAY)`: The first day in the week that contains the
    value to truncate. Weeks begin on `WEEKDAY`. `WEEKDAY` must be one of the
     following: `SUNDAY`, `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`,
     or `SATURDAY`.

  + `ISOWEEK`: The first day in the [ISO 8601 week][ISO-8601-week] that contains
    the value to truncate. The ISO week begins on
    Monday. The first ISO week of each ISO year contains the first Thursday of the
    corresponding Gregorian calendar year.

  + `MONTH`: The first day in the month that contains the
    value to truncate.

  + `QUARTER`: The first day in the quarter that contains the
    value to truncate.

  + `YEAR`: The first day in the year that contains the
    value to truncate.

  + `ISOYEAR`: The first day in the [ISO 8601][ISO-8601] week-numbering year
    that contains the value to truncate. The ISO year is the
    Monday of the first week where Thursday belongs to the corresponding
    Gregorian calendar year.

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[ISO-8601]: https://en.wikipedia.org/wiki/ISO_8601

[ISO-8601-week]: https://en.wikipedia.org/wiki/ISO_week_date

<!-- mdlint on -->

<a id="datetime_trunc_granularity_time"></a>

**Time granularity definitions**

  + `NANOSECOND`: If used, nothing is truncated from the value.

  + `MICROSECOND`: The nearest lesser than or equal microsecond.

  + `MILLISECOND`: The nearest lesser than or equal millisecond.

  + `SECOND`: The nearest lesser than or equal second.

  + `MINUTE`: The nearest lesser than or equal minute.

  + `HOUR`: The nearest lesser than or equal hour.

<a id="datetime_time_zone_parts"></a>

**Time zone part definitions**

+ `MINUTE`
+ `HOUR`
+ `DAY`
+ `WEEK`
+ `WEEK(<WEEKDAY>)`
+ `ISOWEEK`
+ `MONTH`
+ `QUARTER`
+ `YEAR`
+ `ISOYEAR`

**Details**

The resulting value is always rounded to the beginning of `granularity`.

**Return Data Type**

The same data type as the first argument passed into this function.

**Examples**

```googlesql
SELECT
  DATETIME "2008-12-25 15:30:00" as original,
  DATETIME_TRUNC(DATETIME "2008-12-25 15:30:00", DAY) as truncated;

/*----------------------------+------------------------+
 | original                   | truncated              |
 +----------------------------+------------------------+
 | 2008-12-25 15:30:00        | 2008-12-25 00:00:00    |
 +----------------------------+------------------------*/
```

In the following example, the original `DATETIME` falls on a Sunday. Because the
`part` is `WEEK(MONDAY)`, `DATE_TRUNC` returns the `DATETIME` for the
preceding Monday.

```googlesql
SELECT
 datetime AS original,
 DATETIME_TRUNC(datetime, WEEK(MONDAY)) AS truncated
FROM (SELECT DATETIME(TIMESTAMP "2017-11-05 00:00:00+00", "UTC") AS datetime);

/*---------------------+---------------------+
 | original            | truncated           |
 +---------------------+---------------------+
 | 2017-11-05 00:00:00 | 2017-10-30 00:00:00 |
 +---------------------+---------------------*/
```

In the following example, the original `datetime_expression` is in the Gregorian
calendar year 2015. However, `DATETIME_TRUNC` with the `ISOYEAR` date part
truncates the `datetime_expression` to the beginning of the ISO year, not the
Gregorian calendar year. The first Thursday of the 2015 calendar year was
2015-01-01, so the ISO year 2015 begins on the preceding Monday, 2014-12-29.
Therefore the ISO year boundary preceding the `datetime_expression`
2015-06-15 00:00:00 is 2014-12-29.

```googlesql
SELECT
  DATETIME_TRUNC('2015-06-15 00:00:00', ISOYEAR) AS isoyear_boundary,
  EXTRACT(ISOYEAR FROM DATETIME '2015-06-15 00:00:00') AS isoyear_number;

/*---------------------+----------------+
 | isoyear_boundary    | isoyear_number |
 +---------------------+----------------+
 | 2014-12-29 00:00:00 | 2015           |
 +---------------------+----------------*/
```

[datetime-trunc-granularity-date]: #datetime_trunc_granularity_date

[datetime-trunc-granularity-time]: #datetime_trunc_granularity_time

[datetime-time-zone-parts]: #datetime_time_zone_parts

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
