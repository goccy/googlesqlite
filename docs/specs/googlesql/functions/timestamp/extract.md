---
name: EXTRACT
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#extract
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/extract.yaml
---

# EXTRACT

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

## `EXTRACT`

```googlesql
EXTRACT(part FROM timestamp_expression [AT TIME ZONE time_zone])
```

**Description**

Returns a value that corresponds to the specified `part` from
a supplied `timestamp_expression`. This function supports an optional
`time_zone` parameter. See
[Time zone definitions][timestamp-link-to-timezone-definitions] for information
on how to specify a time zone.

Allowed `part` values are:

+ `PICOSECOND`
+ `NANOSECOND`
+ `MICROSECOND`
+ `MILLISECOND`
+ `SECOND`
+ `MINUTE`
+ `HOUR`
+ `DAYOFWEEK`: Returns values in the range [1,7] with Sunday as the first day of
   of the week.
+ `DAY`
+ `DAYOFYEAR`
+ `WEEK`: Returns the week number of the date in the range [0, 53]. Weeks begin
  with Sunday, and dates prior to the first Sunday of the year are in week
  0.
+ `WEEK(<WEEKDAY>)`: Returns the week number of `timestamp_expression` in the
  range [0, 53]. Weeks begin on `WEEKDAY`. `datetime`s prior to the first
  `WEEKDAY` of the year are in week 0. Valid values for `WEEKDAY` are `SUNDAY`,
  `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, and `SATURDAY`.
+ `ISOWEEK`: Returns the [ISO 8601 week][ISO-8601-week]
  number of the `datetime_expression`. `ISOWEEK`s begin on Monday. Return values
  are in the range [1, 53]. The first `ISOWEEK` of each ISO year begins on the
  Monday before the first Thursday of the Gregorian calendar year.
+ `MONTH`
+ `QUARTER`
+ `YEAR`
+ `ISOYEAR`: Returns the [ISO 8601][ISO-8601]
  week-numbering year, which is the Gregorian calendar year containing the
  Thursday of the week to which `date_expression` belongs.
+ `DATE`
+ `DATETIME`
+ `TIME`

Returned values truncate lower order time periods. For example, when extracting
seconds, `EXTRACT` truncates the millisecond and microsecond values.

**Return Data Type**

`INT64`, except in the following cases:

+ If `part` is `DATE`, the function returns a `DATE` object.

**Examples**

In the following example, `EXTRACT` returns a value corresponding to the `DAY`
time part.

```googlesql
SELECT
  EXTRACT(
    DAY
    FROM TIMESTAMP('2008-12-25 05:30:00+00') AT TIME ZONE 'UTC')
    AS the_day_utc,
  EXTRACT(
    DAY
    FROM TIMESTAMP('2008-12-25 05:30:00+00') AT TIME ZONE 'America/Los_Angeles')
    AS the_day_california

/*-------------+--------------------+
 | the_day_utc | the_day_california |
 +-------------+--------------------+
 | 25          | 24                 |
 +-------------+--------------------*/
```

In the following examples, `EXTRACT` returns values corresponding to different
time parts from a column of type `TIMESTAMP`.

```googlesql
SELECT
  EXTRACT(ISOYEAR FROM TIMESTAMP("2005-01-03 12:34:56+00")) AS isoyear,
  EXTRACT(ISOWEEK FROM TIMESTAMP("2005-01-03 12:34:56+00")) AS isoweek,
  EXTRACT(YEAR FROM TIMESTAMP("2005-01-03 12:34:56+00")) AS year,
  EXTRACT(WEEK FROM TIMESTAMP("2005-01-03 12:34:56+00")) AS week

-- Display of results may differ, depending upon the environment and
-- time zone where this query was executed.
/*---------+---------+------+------+
 | isoyear | isoweek | year | week |
 +---------+---------+------+------+
 | 2005    | 1       | 2005 | 1    |
 +---------+---------+------+------*/
```

```googlesql
SELECT
  TIMESTAMP("2007-12-31 12:00:00+00") AS timestamp_value,
  EXTRACT(ISOYEAR FROM TIMESTAMP("2007-12-31 12:00:00+00")) AS isoyear,
  EXTRACT(ISOWEEK FROM TIMESTAMP("2007-12-31 12:00:00+00")) AS isoweek,
  EXTRACT(YEAR FROM TIMESTAMP("2007-12-31 12:00:00+00")) AS year,
  EXTRACT(WEEK FROM TIMESTAMP("2007-12-31 12:00:00+00")) AS week

-- Display of results may differ, depending upon the environment and time zone
-- where this query was executed.
/*---------+---------+------+------+
 | isoyear | isoweek | year | week |
 +---------+---------+------+------+
 | 2008    | 1       | 2007 | 52    |
 +---------+---------+------+------*/
```

```googlesql
SELECT
  TIMESTAMP("2009-01-01 12:00:00+00") AS timestamp_value,
  EXTRACT(ISOYEAR FROM TIMESTAMP("2009-01-01 12:00:00+00")) AS isoyear,
  EXTRACT(ISOWEEK FROM TIMESTAMP("2009-01-01 12:00:00+00")) AS isoweek,
  EXTRACT(YEAR FROM TIMESTAMP("2009-01-01 12:00:00+00")) AS year,
  EXTRACT(WEEK FROM TIMESTAMP("2009-01-01 12:00:00+00")) AS week

-- Display of results may differ, depending upon the environment and time zone
-- where this query was executed.
/*---------+---------+------+------+
 | isoyear | isoweek | year | week |
 +---------+---------+------+------+
 | 2009    | 1       | 2009 | 0    |
 +---------+---------+------+------*/
```

```googlesql
SELECT
  TIMESTAMP("2009-12-31 12:00:00+00") AS timestamp_value,
  EXTRACT(ISOYEAR FROM TIMESTAMP("2009-12-31 12:00:00+00")) AS isoyear,
  EXTRACT(ISOWEEK FROM TIMESTAMP("2009-12-31 12:00:00+00")) AS isoweek,
  EXTRACT(YEAR FROM TIMESTAMP("2009-12-31 12:00:00+00")) AS year,
  EXTRACT(WEEK FROM TIMESTAMP("2009-12-31 12:00:00+00")) AS week

-- Display of results may differ, depending upon the environment and time zone
-- where this query was executed.
/*---------+---------+------+------+
 | isoyear | isoweek | year | week |
 +---------+---------+------+------+
 | 2009    | 53      | 2009 | 52   |
 +---------+---------+------+------*/
```

```googlesql
SELECT
  TIMESTAMP("2017-01-02 12:00:00+00") AS timestamp_value,
  EXTRACT(ISOYEAR FROM TIMESTAMP("2017-01-02 12:00:00+00")) AS isoyear,
  EXTRACT(ISOWEEK FROM TIMESTAMP("2017-01-02 12:00:00+00")) AS isoweek,
  EXTRACT(YEAR FROM TIMESTAMP("2017-01-02 12:00:00+00")) AS year,
  EXTRACT(WEEK FROM TIMESTAMP("2017-01-02 12:00:00+00")) AS week

-- Display of results may differ, depending upon the environment and time zone
-- where this query was executed.
/*---------+---------+------+------+
 | isoyear | isoweek | year | week |
 +---------+---------+------+------+
 | 2017    | 1       | 2017 | 1    |
 +---------+---------+------+------*/
```

```googlesql
SELECT
  TIMESTAMP("2017-05-26 12:00:00+00") AS timestamp_value,
  EXTRACT(ISOYEAR FROM TIMESTAMP("2017-05-26 12:00:00+00")) AS isoyear,
  EXTRACT(ISOWEEK FROM TIMESTAMP("2017-05-26 12:00:00+00")) AS isoweek,
  EXTRACT(YEAR FROM TIMESTAMP("2017-05-26 12:00:00+00")) AS year,
  EXTRACT(WEEK FROM TIMESTAMP("2017-05-26 12:00:00+00")) AS week

-- Display of results may differ, depending upon the environment and time zone
-- where this query was executed.
/*---------+---------+------+------+
 | isoyear | isoweek | year | week |
 +---------+---------+------+------+
 | 2017    | 21      | 2017 | 21   |
 +---------+---------+------+------*/
```

In the following example, `timestamp_expression` falls on a Monday. `EXTRACT`
calculates the first column using weeks that begin on Sunday, and it calculates
the second column using weeks that begin on Monday.

```googlesql
SELECT
  EXTRACT(WEEK(SUNDAY) FROM TIMESTAMP("2017-11-06 00:00:00+00")) AS week_sunday,
  EXTRACT(WEEK(MONDAY) FROM TIMESTAMP("2017-11-06 00:00:00+00")) AS week_monday

-- Display of results may differ, depending upon the environment and time zone
-- where this query was executed.
/*-------------+---------------+
 | week_sunday | week_monday   |
 +-------------+---------------+
 | 45          | 44            |
 +-------------+---------------*/
```

[ISO-8601]: https://en.wikipedia.org/wiki/ISO_8601

[ISO-8601-week]: https://en.wikipedia.org/wiki/ISO_week_date

[timestamp-link-to-timezone-definitions]: #timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
