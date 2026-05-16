---
name: DATETIME
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#datetime
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/datetime.yaml
---

# DATETIME

## Summary

Constructs a `DATETIME` value from individual `INT64` date and time
parts, from a `DATE` (with an optional `TIME`), or from a `TIMESTAMP`
(with an optional time zone).

## Signatures

- `DATETIME(year, month, day, hour, minute, second)`
- `DATETIME(date_expression[, time_expression])`
- `DATETIME(timestamp_expression[, time_zone])`

## Behavior

- Returns a value of type `DATETIME`.
- The six-argument form builds a `DATETIME` from `INT64` year,
  month, day, hour, minute, and second parts.
- The `DATE` form combines a `DATE` with an optional `TIME`; when
  `TIME` is omitted the time component defaults to midnight.
- The `TIMESTAMP` form converts the timestamp to civil date-time
  parts in the supplied time zone.
- When no time zone is supplied to the `TIMESTAMP` form, the
  implementation-defined default time zone is used.

## Examples

```googlesql
SELECT DATETIME(2008, 12, 25, 05, 30, 00) AS datetime_ymdhms;
-- expected: 2008-12-25 05:30:00
```

```googlesql
SELECT DATETIME(TIMESTAMP "2008-12-25 05:30:00+00",
                "America/Los_Angeles") AS datetime_tstz;
-- expected: 2008-12-24 21:30:00
```

## Edge cases

- The default time zone applied to the `TIMESTAMP` form is
  implementation defined, so omitting it can yield different
  civil date-time parts across environments.
- The optional `time_expression` in the `DATE` form and the
  optional `time_zone` in the `TIMESTAMP` form may be omitted,
  but cannot be combined with one another.
- Upstream documentation does not call out explicit NULL or
  out-of-range handling for the part-based form.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/datetime_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATETIME`

```googlesql
1. DATETIME(year, month, day, hour, minute, second)
2. DATETIME(date_expression[, time_expression])
3. DATETIME(timestamp_expression [, time_zone])
```

**Description**

1. Constructs a `DATETIME` object using `INT64` values
   representing the year, month, day, hour, minute, and second.
2. Constructs a `DATETIME` object using a DATE object and an optional `TIME`
   object.
3. Constructs a `DATETIME` object using a `TIMESTAMP` object. It supports an
   optional parameter to
   [specify a time zone][datetime-timezone-definitions].
   If no time zone is specified, the default time zone, which is implementation defined,
   is used.

**Return Data Type**

`DATETIME`

**Example**

```googlesql
SELECT
  DATETIME(2008, 12, 25, 05, 30, 00) as datetime_ymdhms,
  DATETIME(TIMESTAMP "2008-12-25 05:30:00+00", "America/Los_Angeles") as datetime_tstz;

/*---------------------+---------------------+
 | datetime_ymdhms     | datetime_tstz       |
 +---------------------+---------------------+
 | 2008-12-25 05:30:00 | 2008-12-24 21:30:00 |
 +---------------------+---------------------*/
```

[datetime-timezone-definitions]: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
