---
name: FORMAT_TIME
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#format_time
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/format_time.yaml
---

# FORMAT_TIME

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

## `FORMAT_TIME`

```googlesql
FORMAT_TIME(format_string, time_expr)
```

**Description**

Formats a `TIME` value according to the specified format string.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][time-format-elements] to use with `time_expr`.
+   `time_expr`: A `TIME` value that represents the time to format.

**Return Data Type**

`STRING`

**Example**

```googlesql
SELECT FORMAT_TIME("%R", TIME "15:30:00") as formatted_time;

/*----------------+
 | formatted_time |
 +----------------+
 | 15:30          |
 +----------------*/
```

[time-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
