---
name: TIME
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#time
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/time.yaml
---

# TIME

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

## `TIME`

```googlesql
1. TIME(hour, minute, second)
2. TIME(timestamp, [time_zone])
3. TIME(datetime)
```

**Description**

1. Constructs a `TIME` object using `INT64`
   values representing the hour, minute, and second.
2. Constructs a `TIME` object using a `TIMESTAMP` object. It supports an
   optional
   parameter to [specify a time zone][time-link-to-timezone-definitions]. If no
   time zone is specified, the default time zone, which is implementation defined, is
   used.
3. Constructs a `TIME` object using a
   `DATETIME` object.

**Return Data Type**

`TIME`

**Example**

```googlesql
SELECT
  TIME(15, 30, 00) as time_hms,
  TIME(TIMESTAMP "2008-12-25 15:30:00+08", "America/Los_Angeles") as time_tstz;

/*----------+-----------+
 | time_hms | time_tstz |
 +----------+-----------+
 | 15:30:00 | 23:30:00  |
 +----------+-----------*/
```

```googlesql
SELECT TIME(DATETIME "2008-12-25 15:30:00.000000") AS time_dt;

/*----------+
 | time_dt  |
 +----------+
 | 15:30:00 |
 +----------*/
```

[time-link-to-timezone-definitions]: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
