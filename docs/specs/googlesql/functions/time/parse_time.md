---
name: PARSE_TIME
dialect: googlesql
category: functions/time
status: implemented
source_url: docs/third_party/googlesql-docs/time_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/time_functions.md#parse_time
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/time/parse_time.yaml
---

# PARSE_TIME

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

## `PARSE_TIME`

```googlesql
PARSE_TIME(format_string, time_string)
```

**Description**

Converts a `STRING` value to a `TIME` value.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][time-format-elements] to use with `time_string`.
+   `time_string`: A `STRING` value that represents the time to parse.

**Details**

Each element in `time_string` must have a corresponding element in
`format_string`. The location of each element in `format_string` must match the
location of each element in `time_string`.

```googlesql
-- This works because elements on both sides match.
SELECT PARSE_TIME("%I:%M:%S", "07:30:00");

-- This produces an error because the seconds element is in different locations.
SELECT PARSE_TIME("%S:%I:%M", "07:30:00");

-- This produces an error because one of the seconds elements is missing.
SELECT PARSE_TIME("%I:%M", "07:30:00");

-- This works because %T can find all matching elements in time_string.
SELECT PARSE_TIME("%T", "07:30:00");
```

The following additional considerations apply when using the `PARSE_TIME`
function:

+ Unspecified fields. Any unspecified field is initialized from
  `00:00:00.0`. For instance, if `seconds` is unspecified then it
  defaults to `00`, and so on.
+ Whitespace. One or more consecutive white spaces in the format string
  matches zero or more consecutive white spaces in the `TIME` string. In
  addition, leading and trailing white spaces in the `TIME` string are always
  allowed, even if they aren't in the format string.
+ Format precedence. When two (or more) format elements have overlapping
  information, the last one generally overrides any earlier ones.
+ Format divergence. `%p` can be used with `am`, `AM`, `pm`, and `PM`.

**Return Data Type**

`TIME`

**Example**

```googlesql
SELECT PARSE_TIME("%H", "15") as parsed_time;

/*-------------+
 | parsed_time |
 +-------------+
 | 15:00:00    |
 +-------------*/
```

```googlesql
SELECT PARSE_TIME('%I:%M:%S %p', '2:23:38 pm') AS parsed_time;

/*-------------+
 | parsed_time |
 +-------------+
 | 14:23:38    |
 +-------------*/
```

[time-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/time_functions.md`.
