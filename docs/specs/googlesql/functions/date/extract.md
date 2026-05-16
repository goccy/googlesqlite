---
name: EXTRACT
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#extract
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/extract.yaml
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

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `EXTRACT`

```googlesql
EXTRACT(part FROM date_expression)
```

**Description**

Returns the value corresponding to the specified date part. The `part` must
be one of:

+   `DAYOFWEEK`: Returns values in the range [1,7] with Sunday as the first day
    of the week.
+   `DAY`
+   `DAYOFYEAR`
+ `WEEK`: Returns the week number of the date in the range [0, 53]. Weeks begin
  with Sunday, and dates prior to the first Sunday of the year are in week
  0.
+ `WEEK(<WEEKDAY>)`: Returns the week number of the date in the range [0, 53].
  Weeks begin on `WEEKDAY`. Dates prior to
  the first `WEEKDAY` of the year are in week 0. Valid values for `WEEKDAY` are
  `SUNDAY`, `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, and
  `SATURDAY`.
+ `ISOWEEK`: Returns the [ISO 8601 week][ISO-8601-week]
  number of the `date_expression`. `ISOWEEK`s begin on Monday. Return values
  are in the range [1, 53]. The first `ISOWEEK` of each ISO year begins on the
  Monday before the first Thursday of the Gregorian calendar year.
+   `MONTH`
+   `QUARTER`: Returns values in the range [1,4].
+   `YEAR`
+   `ISOYEAR`: Returns the [ISO 8601][ISO-8601]
    week-numbering year, which is the Gregorian calendar year containing the
    Thursday of the week to which `date_expression` belongs.

**Return Data Type**

INT64

**Examples**

In the following example, `EXTRACT` returns a value corresponding to the `DAY`
date part.

```googlesql
SELECT EXTRACT(DAY FROM DATE '2013-12-25') AS the_day;

/*---------+
 | the_day |
 +---------+
 | 25      |
 +---------*/
```

In the following example, `EXTRACT` returns values corresponding to different
date parts from a column of dates near the end of the year.

```googlesql
SELECT
  date,
  EXTRACT(ISOYEAR FROM date) AS isoyear,
  EXTRACT(ISOWEEK FROM date) AS isoweek,
  EXTRACT(YEAR FROM date) AS year,
  EXTRACT(WEEK FROM date) AS week
FROM UNNEST(GENERATE_DATE_ARRAY('2015-12-23', '2016-01-09')) AS date
ORDER BY date;

/*------------+---------+---------+------+------+
 | date       | isoyear | isoweek | year | week |
 +------------+---------+---------+------+------+
 | 2015-12-23 | 2015    | 52      | 2015 | 51   |
 | 2015-12-24 | 2015    | 52      | 2015 | 51   |
 | 2015-12-25 | 2015    | 52      | 2015 | 51   |
 | 2015-12-26 | 2015    | 52      | 2015 | 51   |
 | 2015-12-27 | 2015    | 52      | 2015 | 52   |
 | 2015-12-28 | 2015    | 53      | 2015 | 52   |
 | 2015-12-29 | 2015    | 53      | 2015 | 52   |
 | 2015-12-30 | 2015    | 53      | 2015 | 52   |
 | 2015-12-31 | 2015    | 53      | 2015 | 52   |
 | 2016-01-01 | 2015    | 53      | 2016 | 0    |
 | 2016-01-02 | 2015    | 53      | 2016 | 0    |
 | 2016-01-03 | 2015    | 53      | 2016 | 1    |
 | 2016-01-04 | 2016    | 1       | 2016 | 1    |
 | 2016-01-05 | 2016    | 1       | 2016 | 1    |
 | 2016-01-06 | 2016    | 1       | 2016 | 1    |
 | 2016-01-07 | 2016    | 1       | 2016 | 1    |
 | 2016-01-08 | 2016    | 1       | 2016 | 1    |
 | 2016-01-09 | 2016    | 1       | 2016 | 1    |
 +------------+---------+---------+------+------*/
```

In the following example, `date_expression` falls on a Sunday. `EXTRACT`
calculates the first column using weeks that begin on Sunday, and it calculates
the second column using weeks that begin on Monday.

```googlesql
WITH table AS (SELECT DATE('2017-11-05') AS date)
SELECT
  date,
  EXTRACT(WEEK(SUNDAY) FROM date) AS week_sunday,
  EXTRACT(WEEK(MONDAY) FROM date) AS week_monday FROM table;

/*------------+-------------+-------------+
 | date       | week_sunday | week_monday |
 +------------+-------------+-------------+
 | 2017-11-05 | 45          | 44          |
 +------------+-------------+-------------*/
```

[ISO-8601]: https://en.wikipedia.org/wiki/ISO_8601

[ISO-8601-week]: https://en.wikipedia.org/wiki/ISO_week_date

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
