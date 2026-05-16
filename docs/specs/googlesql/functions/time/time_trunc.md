---
name: TIME_TRUNC
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#time_trunc
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/time_trunc.yaml
---

# TIME_TRUNC

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

## `TIME_TRUNC`

```googlesql
TIME_TRUNC(time_value, time_granularity)
```

**Description**

Truncates a `TIME` value at a particular granularity.

**Definitions**

+ `time_value`: The `TIME` value to truncate.
+ `time_granularity`: The truncation granularity for a `TIME` value.
  [Time granularities][time-trunc-granularity-time] can be used.

<a id="time_trunc_granularity_time"></a>

**Time granularity definitions**

  + `NANOSECOND`: If used, nothing is truncated from the value.

  + `MICROSECOND`: The nearest lesser than or equal microsecond.

  + `MILLISECOND`: The nearest lesser than or equal millisecond.

  + `SECOND`: The nearest lesser than or equal second.

  + `MINUTE`: The nearest lesser than or equal minute.

  + `HOUR`: The nearest lesser than or equal hour.

**Details**

The resulting value is always rounded to the beginning of `granularity`.

**Return Data Type**

`TIME`

**Example**

```googlesql
SELECT
  TIME "15:30:00" as original,
  TIME_TRUNC(TIME "15:30:00", HOUR) as truncated;

/*----------------------------+------------------------+
 | original                   | truncated              |
 +----------------------------+------------------------+
 | 15:30:00                   | 15:00:00               |
 +----------------------------+------------------------*/
```

[time-trunc-granularity-time]: #time_trunc_granularity_time

[time-to-string]: https://github.com/google/googlesql/blob/master/docs/conversion_functions.md#cast

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
