---
name: CURRENT_DATETIME
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#current_datetime
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/current_datetime.yaml
---

# CURRENT_DATETIME

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

## `CURRENT_DATETIME`

```googlesql
CURRENT_DATETIME([time_zone])
```

```googlesql
CURRENT_DATETIME
```

**Description**

Returns the current time as a `DATETIME` object. Parentheses are optional when
called with no arguments.

This function supports an optional `time_zone` parameter.
See [Time zone definitions][datetime-timezone-definitions] for
information on how to specify a time zone.

The current date and time value is set at the start of the query statement that
contains this function. All invocations of `CURRENT_DATETIME()` within a query
statement yield the same value.

**Return Data Type**

`DATETIME`

**Example**

```googlesql
SELECT CURRENT_DATETIME() as now;

/*----------------------------+
 | now                        |
 +----------------------------+
 | 2016-05-19 10:38:47.046465 |
 +----------------------------*/
```

[datetime-timezone-definitions]: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
