---
name: UNIX_SECONDS
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#unix_seconds
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/unix_seconds.yaml
---

# UNIX_SECONDS

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

## `UNIX_SECONDS`

```googlesql
UNIX_SECONDS(timestamp_expression)
```

**Description**

Returns the number of seconds since `1970-01-01 00:00:00 UTC`. Truncates higher
levels of precision by rounding down to the beginning of the second.

**Return Data Type**

`INT64`

**Examples**

```googlesql
SELECT UNIX_SECONDS(TIMESTAMP "2008-12-25 15:30:00+00") AS seconds;

/*------------+
 | seconds    |
 +------------+
 | 1230219000 |
 +------------*/
```

```googlesql
SELECT UNIX_SECONDS(TIMESTAMP "1970-01-01 00:00:01.8+00") AS seconds;

/*------------+
 | seconds    |
 +------------+
 | 1          |
 +------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
