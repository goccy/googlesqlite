---
name: PARSE_DATE
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#parse_date
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/parse_date.yaml
---

# PARSE_DATE

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

## `PARSE_DATE`

```googlesql
PARSE_DATE(format_string, date_string)
```

**Description**

Converts a `STRING` value to a `DATE` value.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][date-format-elements] to use with `date_string`.
+   `date_string`: A `STRING` value that represents the date to parse.

**Details**

Each element in `date_string` must have a corresponding element in
`format_string`. The location of each element in `format_string` must match the
location of each element in `date_string`.

```googlesql
-- This works because elements on both sides match.
SELECT PARSE_DATE('%A %b %e %Y', 'Thursday Dec 25 2008');

-- This produces an error because the year element is in different locations.
SELECT PARSE_DATE('%Y %A %b %e', 'Thursday Dec 25 2008');

-- This produces an error because one of the year elements is missing.
SELECT PARSE_DATE('%A %b %e', 'Thursday Dec 25 2008');

-- This works because %F can find all matching elements in date_string.
SELECT PARSE_DATE('%F', '2000-12-30');
```

The following additional considerations apply when using the `PARSE_DATE`
function:

+ Unspecified fields. Any unspecified field is initialized from `1970-01-01`.
+ Case insensitivity. Names, such as `Monday`, `February`, and so on, are
  case insensitive.
+ Whitespace. One or more consecutive white spaces in the format string
  matches zero or more consecutive white spaces in the date string. In
  addition, leading and trailing white spaces in the date string are always
  allowed, even if they aren't in the format string.
+ Format precedence. When two (or more) format elements have overlapping
  information (for example both `%F` and `%Y` affect the year), the last one
  generally overrides any earlier ones.
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

`DATE`

**Examples**

This example converts a `MM/DD/YY` formatted string to a `DATE` object:

```googlesql
SELECT PARSE_DATE('%x', '12/25/08') AS parsed;

/*------------+
 | parsed     |
 +------------+
 | 2008-12-25 |
 +------------*/
```

This example converts a `YYYYMMDD` formatted string to a `DATE` object:

```googlesql
SELECT PARSE_DATE('%Y%m%d', '20081225') AS parsed;

/*------------+
 | parsed     |
 +------------+
 | 2008-12-25 |
 +------------*/
```

[date-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
