---
name: DATETIME_SUB
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#datetime_sub
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/datetime_sub.yaml
---

# DATETIME_SUB

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

## `DATETIME_SUB`

```googlesql
DATETIME_SUB(datetime_expression, INTERVAL int64_expression part)
```

**Description**

Subtracts `int64_expression` units of `part` from the `DATETIME`.

`DATETIME_SUB` supports the following values for `part`:

+ `NANOSECOND`
+ `MICROSECOND`
+ `MILLISECOND`
+ `SECOND`
+ `MINUTE`
+ `HOUR`
+ `DAY`
+ `WEEK`. Equivalent to 7 `DAY`s.
+ `MONTH`
+ `QUARTER`
+ `YEAR`

Special handling is required for `MONTH`, `QUARTER`, and `YEAR` parts when the
date is at (or near) the last day of the month. If the resulting month has fewer
days than the original `DATETIME`'s day, then the result day is the last day of
the new month.

**Return Data Type**

`DATETIME`

**Example**

```googlesql
SELECT
  DATETIME "2008-12-25 15:30:00" as original_date,
  DATETIME_SUB(DATETIME "2008-12-25 15:30:00", INTERVAL 10 MINUTE) as earlier;

/*-----------------------------+------------------------+
 | original_date               | earlier                |
 +-----------------------------+------------------------+
 | 2008-12-25 15:30:00         | 2008-12-25 15:20:00    |
 +-----------------------------+------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
