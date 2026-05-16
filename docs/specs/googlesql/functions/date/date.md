---
name: DATE
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#date
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/date.yaml
---

# DATE

## Summary

Constructs or extracts a `DATE` value, either from year/month/day integer parts or by extracting the date portion from a `TIMESTAMP` or `DATETIME` expression.

## Signatures

- `DATE(year, month, day)`
- `DATE(timestamp_expression)`
- `DATE(timestamp_expression, time_zone_expression)`
- `DATE(datetime_expression)`

## Behavior

- Returns a value of type `DATE`.
- When called with three `INT64` arguments, constructs a date from `year`, `month`, and `day`.
- When called with a `TIMESTAMP` argument, extracts the date portion of the timestamp.
- When a `time_zone_expression` `STRING` is supplied alongside a `TIMESTAMP`, the timestamp is converted into that time zone before its date portion is extracted.
- When no `time_zone_expression` is supplied with a `TIMESTAMP`, the implementation-defined default time zone is used.
- When called with a `DATETIME` argument, extracts the date portion of the datetime.

## Examples

```googlesql
SELECT
  DATE(2016, 12, 25) AS date_ymd,
  DATE(DATETIME '2016-12-25 23:59:59') AS date_dt,
  DATE(TIMESTAMP '2016-12-25 05:30:00+07', 'America/Los_Angeles') AS date_tstz;
-- expected: date_ymd=2016-12-25, date_dt=2016-12-25, date_tstz=2016-12-24
```

## Edge cases

- The default time zone applied to a `TIMESTAMP` argument without `time_zone_expression` is implementation-defined and may differ between environments.
- Converting a `TIMESTAMP` through a non-UTC `time_zone_expression` can shift the resulting date across the day boundary relative to the UTC date.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATE`

```googlesql
DATE(year, month, day)
```

```googlesql
DATE(timestamp_expression)
```

```googlesql
DATE(timestamp_expression, time_zone_expression)
```

```
DATE(datetime_expression)
```

**Description**

Constructs or extracts a date.

This function supports the following arguments:

+ `year`: The `INT64` value for year.
+ `month`: The `INT64` value for month.
+ `day`: The `INT64` value for day.
+ `timestamp_expression`: A `TIMESTAMP` expression that contains the date.
+ `time_zone_expression`: A `STRING` expression that represents a
  [time zone][date-timezone-definitions]. If no time zone is specified with
  `timestamp_expression`, the default time zone, which is implementation defined, is
  used.
+ `datetime_expression`: A `DATETIME` expression that contains the date.

**Return Data Type**

`DATE`

**Example**

```googlesql
SELECT
  DATE(2016, 12, 25) AS date_ymd,
  DATE(DATETIME '2016-12-25 23:59:59') AS date_dt,
  DATE(TIMESTAMP '2016-12-25 05:30:00+07', 'America/Los_Angeles') AS date_tstz;

/*------------+------------+------------+
 | date_ymd   | date_dt    | date_tstz  |
 +------------+------------+------------+
 | 2016-12-25 | 2016-12-25 | 2016-12-24 |
 +------------+------------+------------*/
```

[date-timezone-definitions]: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
