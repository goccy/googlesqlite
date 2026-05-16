---
name: STRING
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/string.yaml
---

# STRING

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

## `STRING`

```googlesql
STRING(timestamp_expression[, time_zone])
```

**Description**

Converts a timestamp to a string. Supports an optional
parameter to specify a time zone. See
[Time zone definitions][timestamp-link-to-timezone-definitions] for information
on how to specify a time zone.

**Return Data Type**

`STRING`

**Example**

```googlesql
SELECT STRING(TIMESTAMP "2008-12-25 15:30:00+00", "UTC") AS string;

/*-------------------------------+
 | string                        |
 +-------------------------------+
 | 2008-12-25 15:30:00+00        |
 +-------------------------------*/
```

[timestamp-link-to-timezone-definitions]: #timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
