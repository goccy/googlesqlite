---
name: UNIX_MICROS
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#unix_micros
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/unix_micros.yaml
---

# UNIX_MICROS

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

## `UNIX_MICROS`

```googlesql
UNIX_MICROS(timestamp_expression)
```

**Description**

Returns the number of microseconds since `1970-01-01 00:00:00 UTC`.
Truncates higher levels of precision by
rounding down to the beginning of the microsecond.

**Return Data Type**

`INT64`

**Examples**

```googlesql
SELECT UNIX_MICROS(TIMESTAMP "2008-12-25 15:30:00+00") AS micros;

/*------------------+
 | micros           |
 +------------------+
 | 1230219000000000 |
 +------------------*/
```

```googlesql
SELECT UNIX_MICROS(TIMESTAMP "1970-01-01 00:00:00.0000018+00") AS micros;

/*------------------+
 | micros           |
 +------------------+
 | 1                |
 +------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
