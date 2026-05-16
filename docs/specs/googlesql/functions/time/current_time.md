---
name: CURRENT_TIME
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#current_time
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/current_time.yaml
---

# CURRENT_TIME

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

Verbatim copy from `docs/third_party/googlesql-docs/time_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `CURRENT_TIME`

```googlesql
CURRENT_TIME([time_zone])
```

```googlesql
CURRENT_TIME
```

**Description**

Returns the current time as a `TIME` object. Parentheses are optional when
called with no arguments.

This function supports an optional `time_zone` parameter.
See [Time zone definitions][time-link-to-timezone-definitions] for information
on how to specify a time zone.

The current time value is set at the start of the query statement that contains
this function. All invocations of `CURRENT_TIME()` within a query statement
yield the same value.

**Return Data Type**

`TIME`

**Example**

```googlesql
SELECT CURRENT_TIME() as now;

/*----------------------------+
 | now                        |
 +----------------------------+
 | 15:31:38.776361            |
 +----------------------------*/
```

[time-link-to-timezone-definitions]: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
