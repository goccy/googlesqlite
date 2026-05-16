---
name: TIMESTAMP_ADD
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timestamp_add
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/timestamp_add.yaml
---

# TIMESTAMP_ADD

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

Verbatim copy from `docs/third_party/googlesql-docs/timestamp_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TIMESTAMP_ADD`

```googlesql
TIMESTAMP_ADD(timestamp_expression, INTERVAL int64_expression date_part)
```

**Description**

Adds `int64_expression` units of `date_part` to the timestamp, independent of
any time zone.

`TIMESTAMP_ADD` supports the following values for `date_part`:

+ `PICOSECOND`
+ `NANOSECOND`
+ `MICROSECOND`
+ `MILLISECOND`
+ `SECOND`
+ `MINUTE`
+ `HOUR`. Equivalent to 60 `MINUTE` parts.
+ `DAY`. Equivalent to 24 `HOUR` parts.

**Return Data Types**

`TIMESTAMP`

**Example**

```googlesql
SELECT
  TIMESTAMP("2008-12-25 15:30:00+00") AS original,
  TIMESTAMP_ADD(TIMESTAMP "2008-12-25 15:30:00+00", INTERVAL 10 MINUTE) AS later;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+---------------------------------------------+
 | original                                    | later                                       |
 +---------------------------------------------+---------------------------------------------+
 | 2008-12-25 07:30:00.000 America/Los_Angeles | 2008-12-25 07:40:00.000 America/Los_Angeles |
 +---------------------------------------------+---------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
