---
name: TIME_DIFF
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#time_diff
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/time_diff.yaml
---

# TIME_DIFF

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

## `TIME_DIFF`

```googlesql
TIME_DIFF(end_time, start_time, granularity)
```

**Description**

Gets the number of unit boundaries between two `TIME` values (`end_time` -
`start_time`) at a particular time granularity.

**Definitions**

+   `start_time`: The starting `TIME` value.
+   `end_time`: The ending `TIME` value.
+   `granularity`: The time part that represents the granularity. If
    you passed in `TIME` values for the first arguments, `granularity` can
    be:

    
    + `NANOSECOND`
    + `MICROSECOND`
    + `MILLISECOND`
    + `SECOND`
    + `MINUTE`
    + `HOUR`

**Details**

If `end_time` is earlier than `start_time`, the output is negative.
Produces an error if the computation overflows, such as if the difference
in nanoseconds
between the two `TIME` values overflows.

Note: The behavior of the this function follows the type of arguments passed in.
For example, `TIME_DIFF(TIMESTAMP, TIMESTAMP, PART)`
behaves like `TIMESTAMP_DIFF(TIMESTAMP, TIMESTAMP, PART)`.

**Return Data Type**

`INT64`

**Example**

```googlesql
SELECT
  TIME "15:30:00" as first_time,
  TIME "14:35:00" as second_time,
  TIME_DIFF(TIME "15:30:00", TIME "14:35:00", MINUTE) as difference;

/*----------------------------+------------------------+------------------------+
 | first_time                 | second_time            | difference             |
 +----------------------------+------------------------+------------------------+
 | 15:30:00                   | 14:35:00               | 55                     |
 +----------------------------+------------------------+------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
