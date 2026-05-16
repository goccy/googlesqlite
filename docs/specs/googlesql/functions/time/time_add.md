---
name: TIME_ADD
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#time_add
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/time_add.yaml
---

# TIME_ADD

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

## `TIME_ADD`

```googlesql
TIME_ADD(time_expression, INTERVAL int64_expression part)
```

**Description**

Adds `int64_expression` units of `part` to the `TIME` object.

`TIME_ADD` supports the following values for `part`:

+ `NANOSECOND`
+ `MICROSECOND`
+ `MILLISECOND`
+ `SECOND`
+ `MINUTE`
+ `HOUR`

This function automatically adjusts when values fall outside of the 00:00:00 to
24:00:00 boundary. For example, if you add an hour to `23:30:00`, the returned
value is `00:30:00`.

**Return Data Types**

`TIME`

**Example**

```googlesql
SELECT
  TIME "15:30:00" as original_time,
  TIME_ADD(TIME "15:30:00", INTERVAL 10 MINUTE) as later;

/*-----------------------------+------------------------+
 | original_time               | later                  |
 +-----------------------------+------------------------+
 | 15:30:00                    | 15:40:00               |
 +-----------------------------+------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
