---
name: TIMESTAMP_TRUNC
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timestamp_trunc
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/timestamp_trunc.yaml
---

# TIMESTAMP_TRUNC

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

Verbatim copy from `docs/third_party/googlesql-docs/timestamp_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TIMESTAMP_TRUNC`

```googlesql
TIMESTAMP_TRUNC(timestamp_value, timestamp_granularity[, time_zone])
```

```googlesql
TIMESTAMP_TRUNC(datetime_value, datetime_granularity)
```

**Description**

Truncates a `TIMESTAMP` or `DATETIME` value at a particular granularity.

**Definitions**

+ `timestamp_value`: A `TIMESTAMP` value to truncate.
+ `timestamp_granularity`: The truncation granularity for a `TIMESTAMP` value.
  [Date granularities][timestamp-trunc-granularity-date] and
  [time granularities][timestamp-trunc-granularity-time] can be used.
+ `time_zone`: A time zone to use with the `TIMESTAMP` value.
  [Time zone parts][timestamp-time-zone-parts] can be used.
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
+ `datetime_value`: A `DATETIME` value to truncate.
+ `datetime_granularity`: The truncation granularity for a `DATETIME` value.
  [Date granularities][timestamp-trunc-granularity-date] and
  [time granularities][timestamp-trunc-granularity-time] can be used.

<a id="timestamp_trunc_granularity_date"></a>

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

<a id="timestamp_trunc_granularity_time"></a>

**Time granularity definitions**

  + `PICOSECOND`: If used, nothing is truncated from the value.

  + `NANOSECOND`: If used, nothing is truncated from the value.

  + `MICROSECOND`: The nearest lesser than or equal microsecond.

  + `MILLISECOND`: The nearest lesser than or equal millisecond.

  + `SECOND`: The nearest lesser than or equal second.

  + `MINUTE`: The nearest lesser than or equal minute.

  + `HOUR`: The nearest lesser than or equal hour.

<a id="timestamp_time_zone_parts"></a>

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
  TIMESTAMP_TRUNC(TIMESTAMP "2008-12-25 15:30:00+00", DAY, "UTC") AS utc,
  TIMESTAMP_TRUNC(TIMESTAMP "2008-12-25 15:30:00+00", DAY, "America/Los_Angeles") AS la;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+---------------------------------------------+
 | utc                                         | la                                          |
 +---------------------------------------------+---------------------------------------------+
 | 2008-12-24 16:00:00.000 America/Los_Angeles | 2008-12-25 00:00:00.000 America/Los_Angeles |
 +---------------------------------------------+---------------------------------------------*/
```

In the following example, `timestamp_expression` has a time zone offset of +12.
The first column shows the `timestamp_expression` in UTC time. The second
column shows the output of `TIMESTAMP_TRUNC` using weeks that start on Monday.
Because the `timestamp_expression` falls on a Sunday in UTC, `TIMESTAMP_TRUNC`
truncates it to the preceding Monday. The third column shows the same function
with the optional [Time zone definition][timestamp-link-to-timezone-definitions]
argument 'Pacific/Auckland'. Here, the function truncates the
`timestamp_expression` using New Zealand Daylight Time, where it falls on a
Monday.

```googlesql
SELECT
  timestamp_value AS timestamp_value,
  TIMESTAMP_TRUNC(timestamp_value, WEEK(MONDAY), "UTC") AS utc_truncated,
  TIMESTAMP_TRUNC(timestamp_value, WEEK(MONDAY), "Pacific/Auckland") AS nzdt_truncated
FROM (SELECT TIMESTAMP("2017-11-06 00:00:00+12") AS timestamp_value);

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+---------------------------------------------+---------------------------------------------+
 | timestamp_value                             | utc_truncated                               | nzdt_truncated                              |
 +---------------------------------------------+---------------------------------------------+---------------------------------------------+
 | 2017-11-05 04:00:00.000 America/Los_Angeles | 2017-10-29 17:00:00.000 America/Los_Angeles | 2017-11-05 03:00:00.000 America/Los_Angeles |
 +---------------------------------------------+---------------------------------------------+---------------------------------------------*/
```

In the following example, the original `timestamp_expression` is in the
Gregorian calendar year 2015. However, `TIMESTAMP_TRUNC` with the `ISOYEAR` date
part truncates the `timestamp_expression` to the beginning of the ISO year, not
the Gregorian calendar year. The first Thursday of the 2015 calendar year was
2015-01-01, so the ISO year 2015 begins on the preceding Monday, 2014-12-29.
Therefore the ISO year boundary preceding the `timestamp_expression`
2015-06-15 00:00:00+00 is 2014-12-29.

```googlesql
SELECT
  TIMESTAMP_TRUNC("2015-06-15 00:00:00+00", ISOYEAR) AS isoyear_boundary,
  EXTRACT(ISOYEAR FROM TIMESTAMP "2015-06-15 00:00:00+00") AS isoyear_number;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+----------------+
 | isoyear_boundary                            | isoyear_number |
 +---------------------------------------------+----------------+
 | 2014-12-29 00:00:00.000 America/Los_Angeles | 2015           |
 +---------------------------------------------+----------------*/
```

[timestamp-link-to-timezone-definitions]: #timezone_definitions

[timestamp-trunc-granularity-date]: #timestamp_trunc_granularity_date

[timestamp-trunc-granularity-time]: #timestamp_trunc_granularity_time

[timestamp-time-zone-parts]: #timestamp_time_zone_parts

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
