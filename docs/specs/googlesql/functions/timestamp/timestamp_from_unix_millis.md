---
name: TIMESTAMP_FROM_UNIX_MILLIS
dialect: googlesql
category: functions/timestamp
status: implemented
notes: |
  Analyzer accepts the signature but the runtime UDF is missing; equivalent to TIMESTAMP_SECONDS / _MILLIS / _MICROS already shipped.  will alias.
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#timestamp_from_unix_millis
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/timestamp_from_unix_millis.yaml
---

# TIMESTAMP_FROM_UNIX_MILLIS

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

## `TIMESTAMP_FROM_UNIX_MILLIS`

```googlesql
TIMESTAMP_FROM_UNIX_MILLIS(int64_expression)
```

```googlesql
TIMESTAMP_FROM_UNIX_MILLIS(timestamp_expression)
```

**Description**

Interprets `int64_expression` as the number of milliseconds since
1970-01-01 00:00:00 UTC and returns a timestamp. If a timestamp is passed in,
the same timestamp is returned.

**Return Data Type**

`TIMESTAMP`

**Example**

```googlesql
SELECT TIMESTAMP_FROM_UNIX_MILLIS(1230219000000) AS timestamp_value;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*------------------------+
 | timestamp_value        |
 +------------------------+
 | 2008-12-25 15:30:00+00 |
 +------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
