---
name: DATE_FROM_UNIX_DATE
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#date_from_unix_date
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/date_from_unix_date.yaml
---

# DATE_FROM_UNIX_DATE

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

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `DATE_FROM_UNIX_DATE`

```googlesql
DATE_FROM_UNIX_DATE(int64_expression)
```

**Description**

Interprets `int64_expression` as the number of days since 1970-01-01.

**Return Data Type**

DATE

**Example**

```googlesql
SELECT DATE_FROM_UNIX_DATE(14238) AS date_from_epoch;

/*-----------------+
 | date_from_epoch |
 +-----------------+
 | 2008-12-25      |
 +-----------------+*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
