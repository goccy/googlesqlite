---
name: TIMESTAMP_DIFF
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timestamp_diff
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/timestamp_diff.yaml
---

# TIMESTAMP_DIFF

## Summary

Gets the number of unit boundaries between two `TIMESTAMP` values
(`end_timestamp` - `start_timestamp`) at a particular time granularity.

## Signatures

- `TIMESTAMP_DIFF(end_timestamp, start_timestamp, granularity)`

## Behavior

- Returns `INT64`.
- Computes `end_timestamp - start_timestamp` and reports the count of
  unit boundaries crossed at the given `granularity`.
- Supported `granularity` values for `TIMESTAMP` arguments are
  `PICOSECOND`, `NANOSECOND`, `MICROSECOND`, `MILLISECOND`, `SECOND`,
  `MINUTE`, `HOUR` (equivalent to 60 `MINUTE`s), and `DAY` (equivalent
  to 24 `HOUR`s).
- Only whole intervals of the specified `granularity` are counted;
  fractional remainders are discarded.
- Behavior follows the type of the arguments passed in; for example,
  `TIMESTAMP_DIFF(DATE, DATE, PART)` behaves like
  `DATE_DIFF(DATE, DATE, PART)`.

## Examples

```googlesql
SELECT TIMESTAMP_DIFF(TIMESTAMP "2010-07-07 10:20:00+00", TIMESTAMP "2008-12-25 15:30:00+00", HOUR) AS hours;
-- expected hours = 13410
```

```googlesql
SELECT TIMESTAMP_DIFF(TIMESTAMP "2018-08-14", TIMESTAMP "2018-10-14", DAY) AS negative_diff;
-- expected negative_diff = -61
```

```googlesql
SELECT TIMESTAMP_DIFF("2001-02-01 01:00:00", "2001-02-01 00:00:01", HOUR) AS diff;
-- expected diff = 0
```

## Edge cases

- If `end_timestamp` is earlier than `start_timestamp`, the result is
  negative.
- Produces an error if the computation overflows, such as when the
  difference in nanoseconds between the two `TIMESTAMP` values
  overflows `INT64`.
- Sub-`granularity` differences round toward zero, so a one-hour gap
  measured one second short returns `0` for `HOUR`.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/timestamp_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TIMESTAMP_DIFF`

```googlesql
TIMESTAMP_DIFF(end_timestamp, start_timestamp, granularity)
```

**Description**

Gets the number of unit boundaries between two `TIMESTAMP` values
(`end_timestamp` - `start_timestamp`) at a particular time granularity.

**Definitions**

+   `start_timestamp`: The starting `TIMESTAMP` value.
+   `end_timestamp`: The ending `TIMESTAMP` value.
+   `granularity`: The timestamp part that represents the granularity. If
    you passed in `TIMESTAMP` values for the first arguments, `granularity` can
    be:

    
    + `PICOSECOND`
    + `NANOSECOND`
    + `MICROSECOND`
    + `MILLISECOND`
    + `SECOND`
    + `MINUTE`
    + `HOUR`. Equivalent to 60 `MINUTE`s.
    + `DAY`. Equivalent to 24 `HOUR`s.

**Details**

If `end_timestamp` is earlier than `start_timestamp`, the output is negative.
Produces an error if the computation overflows, such as if the difference
in nanoseconds
between the two `TIMESTAMP` values overflows.

Note: The behavior of the this function follows the type of arguments passed in.
For example, `TIMESTAMP_DIFF(DATE, DATE, PART)`
behaves like `DATE_DIFF(DATE, DATE, PART)`.

**Return Data Type**

`INT64`

**Example**

```googlesql
SELECT
  TIMESTAMP("2010-07-07 10:20:00+00") AS later_timestamp,
  TIMESTAMP("2008-12-25 15:30:00+00") AS earlier_timestamp,
  TIMESTAMP_DIFF(TIMESTAMP "2010-07-07 10:20:00+00", TIMESTAMP "2008-12-25 15:30:00+00", HOUR) AS hours;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+---------------------------------------------+-------+
 | later_timestamp                             | earlier_timestamp                           | hours |
 +---------------------------------------------+---------------------------------------------+-------+
 | 2010-07-07 03:20:00.000 America/Los_Angeles | 2008-12-25 07:30:00.000 America/Los_Angeles | 13410 |
 +---------------------------------------------+---------------------------------------------+-------*/
```

In the following example, the first timestamp occurs before the
second timestamp, resulting in a negative output.

```googlesql
SELECT TIMESTAMP_DIFF(TIMESTAMP "2018-08-14", TIMESTAMP "2018-10-14", DAY) AS negative_diff;

/*---------------+
 | negative_diff |
 +---------------+
 | -61           |
 +---------------*/
```

In this example, the result is 0 because only the number of whole specified
`HOUR` intervals are included.

```googlesql
SELECT TIMESTAMP_DIFF("2001-02-01 01:00:00", "2001-02-01 00:00:01", HOUR) AS diff;

/*---------------+
 | diff          |
 +---------------+
 | 0             |
 +---------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
