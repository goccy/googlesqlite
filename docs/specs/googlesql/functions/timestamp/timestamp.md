---
name: TIMESTAMP
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timestamp
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/timestamp.yaml
---

# TIMESTAMP

## Summary

Constructs a `TIMESTAMP` value from a string, `DATE`, or `DATETIME` expression, with an optional time zone argument used when the input itself is not time-zone-aware.

## Signatures

- `TIMESTAMP(string_expression[, time_zone])`
- `TIMESTAMP(date_expression[, time_zone])`
- `TIMESTAMP(datetime_expression[, time_zone])`

## Behavior

- Returns a `TIMESTAMP`.
- For a string input, the string must contain a timestamp literal; the optional `time_zone` argument applies when the literal does not itself encode a time zone.
- If the string literal already includes a time zone, an explicit `time_zone` argument must not be supplied.
- For a `DATE` input, the result is the earliest timestamp that falls within the given date (i.e. midnight at the start of the day) in the given or default time zone.
- For a `DATETIME` input, the civil datetime is interpreted in the given or default time zone and converted to an absolute timestamp.
- When no `time_zone` argument is provided, the implementation-defined default time zone is used.

## Examples

```googlesql
SELECT TIMESTAMP("2008-12-25 15:30:00+00") AS timestamp_str;
-- expected: 2008-12-25 07:30:00.000 America/Los_Angeles
```

```googlesql
SELECT TIMESTAMP("2008-12-25 15:30:00", "America/Los_Angeles") AS timestamp_str;
-- expected: 2008-12-25 15:30:00.000 America/Los_Angeles
```

```googlesql
SELECT TIMESTAMP(DATE "2008-12-25") AS timestamp_date;
-- expected: 2008-12-25 00:00:00.000 America/Los_Angeles
```

## Edge cases

- Supplying an explicit `time_zone` argument together with a string literal that already encodes a time zone is not allowed.
- The default time zone used when `time_zone` is omitted is implementation defined, so identical inputs may yield different absolute timestamps across environments.
- Display of resulting timestamps depends on the session/environment time zone and may differ from the stored absolute instant.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/timestamp_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TIMESTAMP`

```googlesql
TIMESTAMP(string_expression[, time_zone])
TIMESTAMP(date_expression[, time_zone])
TIMESTAMP(datetime_expression[, time_zone])
```

**Description**

+  `string_expression[, time_zone]`: Converts a string to a
   timestamp. `string_expression` must include a
   timestamp literal.
   If `string_expression` includes a time zone in the timestamp literal,
   don't include an explicit `time_zone`
   argument.
+  `date_expression[, time_zone]`: Converts a date to a timestamp.
   The value returned is the earliest timestamp that falls within
   the given date.
+  `datetime_expression[, time_zone]`: Converts a
   datetime to a timestamp.

This function supports an optional
parameter to [specify a time zone][timestamp-link-to-timezone-definitions]. If
no time zone is specified, the default time zone, which is implementation defined,
is used.

**Return Data Type**

`TIMESTAMP`

**Examples**

```googlesql
SELECT TIMESTAMP("2008-12-25 15:30:00+00") AS timestamp_str;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+
 | timestamp_str                               |
 +---------------------------------------------+
 | 2008-12-25 07:30:00.000 America/Los_Angeles |
 +---------------------------------------------*/
```

```googlesql
SELECT TIMESTAMP("2008-12-25 15:30:00", "America/Los_Angeles") AS timestamp_str;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+
 | timestamp_str                               |
 +---------------------------------------------+
 | 2008-12-25 15:30:00.000 America/Los_Angeles |
 +---------------------------------------------*/
```

```googlesql
SELECT TIMESTAMP("2008-12-25 15:30:00 UTC") AS timestamp_str;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+
 | timestamp_str                               |
 +---------------------------------------------+
 | 2008-12-25 07:30:00.000 America/Los_Angeles |
 +---------------------------------------------*/
```

```googlesql
SELECT TIMESTAMP(DATETIME "2008-12-25 15:30:00") AS timestamp_datetime;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+
 | timestamp_datetime                          |
 +---------------------------------------------+
 | 2008-12-25 15:30:00.000 America/Los_Angeles |
 +---------------------------------------------*/
```

```googlesql
SELECT TIMESTAMP(DATE "2008-12-25") AS timestamp_date;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+
 | timestamp_date                              |
 +---------------------------------------------+
 | 2008-12-25 00:00:00.000 America/Los_Angeles |
 +---------------------------------------------*/
```

[timestamp-link-to-timezone-definitions]: #timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
