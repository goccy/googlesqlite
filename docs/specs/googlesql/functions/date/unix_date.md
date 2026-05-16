---
name: UNIX_DATE
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#unix_date
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/unix_date.yaml
---

# UNIX_DATE

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

## `UNIX_DATE`

```googlesql
UNIX_DATE(date_expression)
```

**Description**

Returns the number of days since `1970-01-01`.

**Return Data Type**

INT64

**Example**

```googlesql
SELECT UNIX_DATE(DATE '2008-12-25') AS days_from_epoch;

/*-----------------+
 | days_from_epoch |
 +-----------------+
 | 14238           |
 +-----------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
