---
name: PARSE_TIMESTAMP
dialect: googlesql
category: functions/timestamp
status: implemented
source_url: docs/third_party/googlesql-docs/timestamp_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/timestamp_functions.md#parse_timestamp
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/timestamp/parse_timestamp.yaml
---

# PARSE_TIMESTAMP

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

## `PARSE_TIMESTAMP`

```googlesql
PARSE_TIMESTAMP(format_string, timestamp_string[, time_zone])
```

**Description**

Converts a `STRING` value to a `TIMESTAMP` value.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][timestamp-format-elements] to use with `timestamp_string`.
+   `timestamp_string`: A `STRING` value that represents the timestamp to parse.
+   `time_zone`: A `STRING` value that represents a time zone. For more
    information about how to use a time zone with a timestamp, see
    [Time zone definitions][timestamp-link-to-timezone-definitions].

**Details**

Each element in `timestamp_string` must have a corresponding element in
`format_string`. The location of each element in `format_string` must match the
location of each element in `timestamp_string`.

```googlesql
-- This works because elements on both sides match.
SELECT PARSE_TIMESTAMP("%a %b %e %I:%M:%S %Y", "Thu Dec 25 07:30:00 2008");

-- This produces an error because the year element is in different locations.
SELECT PARSE_TIMESTAMP("%a %b %e %Y %I:%M:%S", "Thu Dec 25 07:30:00 2008");

-- This produces an error because one of the year elements is missing.
SELECT PARSE_TIMESTAMP("%a %b %e %I:%M:%S", "Thu Dec 25 07:30:00 2008");

-- This works because %c can find all matching elements in timestamp_string.
SELECT PARSE_TIMESTAMP("%c", "Thu Dec 25 07:30:00 2008");
```

The following additional considerations apply when using the `PARSE_TIMESTAMP`
function:

+ Unspecified fields. Any unspecified field is initialized from `1970-01-01
  00:00:00.0`. This initialization value uses the time zone specified by the
  function's time zone argument, if present. If not, the initialization value
  uses the default time zone, which is implementation defined. For instance, if the year
  is unspecified then it defaults to `1970`, and so on.
+ Case insensitivity. Names, such as `Monday`, `February`, and so on, are
  case insensitive.
+ Whitespace. One or more consecutive white spaces in the format string
  matches zero or more consecutive white spaces in the timestamp string. In
  addition, leading and trailing white spaces in the timestamp string are always
  allowed, even if they aren't in the format string.
+ Format precedence. When two (or more) format elements have overlapping
  information (for example both `%F` and `%Y` affect the year), the last one
  generally overrides any earlier ones, with some exceptions (see the
  descriptions of `%s`, `%C`, and `%y`).
+ Format divergence. `%p` can be used with `am`, `AM`, `pm`, and `PM`.
+   Mixed ISO and non-ISO elements. The ISO format elements are `%G`, `%g`,
    `%J`, and `%V`. When these ISO elements are used together with other non-ISO
    elements, the ISO elements are ignored, resulting in different values. For
    example, the function arguments `('%g %J', '8405')` return a value with the
    year `1984`, whereas the arguments `('%g %j', '8405')` return a value with
    the year `1970` because the ISO element `%g` is ignored.
+   Numeric values after `%G` input values. Any input string value that
    corresponds to the `%G` format element requires a whitespace or non-digit
    character as a separator from numeric values that follow. This is a known
    issue in GoogleSQL. For example, the function arguments `('%G
    %V','2020 50')` or `('%G-%V','2020-50')` work, but not `('%G%V','202050')`.
    For input values before the corresponding `%G` value, no separator is
    needed. For example, the arguments `('%V%G','502020')` work. The separator
    after the `%G` values identifies the end of the specified ISO year value so
    that the function can parse properly.

**Return Data Type**

`TIMESTAMP`

**Example**

```googlesql
SELECT PARSE_TIMESTAMP("%c", "Thu Dec 25 07:30:00 2008") AS parsed;

-- Display of results may differ, depending upon the environment and time zone where this query was executed.
/*---------------------------------------------+
 | parsed                                      |
 +---------------------------------------------+
 | 2008-12-25 07:30:00.000 America/Los_Angeles |
 +---------------------------------------------*/
```

[timestamp-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

[timestamp-link-to-timezone-definitions]: #timezone_definitions

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/timestamp_functions.md`.
