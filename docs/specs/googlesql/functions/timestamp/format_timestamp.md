---
name: FORMAT_TIMESTAMP
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#format_timestamp
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/format_timestamp.yaml
---

# FORMAT_TIMESTAMP

## Summary

Formats a `TIMESTAMP` value as a `STRING` according to the supplied format string, optionally interpreted in a given time zone.

## Signatures

- `FORMAT_TIMESTAMP(format_string, timestamp_expr[, time_zone])`

## Behavior

- Returns `STRING`.
- `format_string` is a `STRING` containing date/time format elements applied to `timestamp_expr`.
- `timestamp_expr` is the `TIMESTAMP` value being formatted.
- `time_zone`, when provided, is a `STRING` naming the time zone used to render the timestamp; otherwise the default time zone applies.
- Format elements follow the googlesql date/time format element set.

## Examples

```googlesql
SELECT FORMAT_TIMESTAMP("%c", TIMESTAMP "2050-12-25 15:30:55+00", "UTC") AS formatted;
-- expected: Sun Dec 25 15:30:55 2050
```

```googlesql
SELECT FORMAT_TIMESTAMP("%b-%d-%Y", TIMESTAMP "2050-12-25 15:30:55+00") AS formatted;
-- expected: Dec-25-2050
```

```googlesql
SELECT FORMAT_TIMESTAMP("%Y-%m-%dT%H:%M:%S%Z", TIMESTAMP "2050-12-25 15:30:55", "UTC") AS formatted;
-- expected: 2050-12-25T15:30:55UTC
```

## Edge cases

- Output depends on the chosen `time_zone`; omitting it falls back to the session/default time zone, which can shift rendered date and time fields.
- Format elements not supported by the upstream specification are not guaranteed to produce defined output.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/timestamp_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `FORMAT_TIMESTAMP`

```googlesql
FORMAT_TIMESTAMP(format_string, timestamp_expr[, time_zone])
```

**Description**

Formats a `TIMESTAMP` value according to the specified format string.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][timestamp-format-elements] to use with
    `timestamp_expr`.
+   `timestamp_expr`: A `TIMESTAMP` value that represents the timestamp to format.
+   `time_zone`: A `STRING` value that represents a time zone. For more
    information about how to use a time zone with a timestamp, see
    [Time zone definitions][timestamp-link-to-timezone-definitions].

**Return Data Type**

`STRING`

**Examples**

```googlesql
SELECT FORMAT_TIMESTAMP("%c", TIMESTAMP "2050-12-25 15:30:55+00", "UTC")
  AS formatted;

/*--------------------------+
 | formatted                |
 +--------------------------+
 | Sun Dec 25 15:30:55 2050 |
 +--------------------------*/
```

```googlesql
SELECT FORMAT_TIMESTAMP("%b-%d-%Y", TIMESTAMP "2050-12-25 15:30:55+00")
  AS formatted;

/*-------------+
 | formatted   |
 +-------------+
 | Dec-25-2050 |
 +-------------*/
```

```googlesql
SELECT FORMAT_TIMESTAMP("%b %Y", TIMESTAMP "2050-12-25 15:30:55+00")
  AS formatted;

/*-------------+
 | formatted   |
 +-------------+
 | Dec 2050    |
 +-------------*/
```

```googlesql
SELECT FORMAT_TIMESTAMP("%Y-%m-%dT%H:%M:%S%Z", TIMESTAMP "2050-12-25 15:30:55", "UTC")
  AS formatted;

/*+-----------------------+
 |       formatted        |
 +------------------------+
 | 2050-12-25T15:30:55UTC |
 +------------------------*/
```

[timestamp-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

[timestamp-link-to-timezone-definitions]: #timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
