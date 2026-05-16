---
name: PARSE_DATETIME
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#parse_datetime
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/parse_datetime.yaml
---

# PARSE_DATETIME

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

Verbatim copy from `docs/third_party/googlesql-docs/datetime_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PARSE_DATETIME`

```googlesql
PARSE_DATETIME(format_string, datetime_string)
```

**Description**

Converts a `STRING` value to a `DATETIME` value.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][datetime-format-elements] to use with `datetime_string`.
+   `datetime_string`: A `STRING` value that represents the date and time to
    parse.

**Details**

Each element in `datetime_string` must have a corresponding element in
`format_string`. The location of each element in `format_string` must match the
location of each element in `datetime_string`.

```googlesql
-- This works because elements on both sides match.
SELECT PARSE_DATETIME("%a %b %e %I:%M:%S %Y", "Thu Dec 25 07:30:00 2008");

-- This produces an error because the year element is in different locations.
SELECT PARSE_DATETIME("%a %b %e %Y %I:%M:%S", "Thu Dec 25 07:30:00 2008");

-- This produces an error because one of the year elements is missing.
SELECT PARSE_DATETIME("%a %b %e %I:%M:%S", "Thu Dec 25 07:30:00 2008");

-- This works because %c can find all matching elements in datetime_string.
SELECT PARSE_DATETIME("%c", "Thu Dec 25 07:30:00 2008");
```

The following additional considerations apply when using the `PARSE_DATETIME`
function:

+ Unspecified fields. Any unspecified field is initialized from
  `1970-01-01 00:00:00.0`. For example, if the year is unspecified then it
  defaults to `1970`.
+ Case insensitivity. Names, such as `Monday` and `February`,
  are case insensitive.
+ Whitespace. One or more consecutive white spaces in the format string
  matches zero or more consecutive white spaces in the
  `DATETIME` string. Leading and trailing
  white spaces in the `DATETIME` string are always
  allowed, even if they aren't in the format string.
+ Format precedence. When two or more format elements have overlapping
  information, the last one generally overrides any earlier ones, with some
  exceptions. For example, both `%F` and `%Y` affect the year, so the earlier
  element overrides the later. See the descriptions
  of `%s`, `%C`, and `%y` in
  [Supported Format Elements For DATETIME][datetime-format-elements].
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

`DATETIME`

**Examples**

The following examples parse a `STRING` literal as a
`DATETIME`.

```googlesql
SELECT PARSE_DATETIME('%Y-%m-%d %H:%M:%S', '1998-10-18 13:45:55') AS datetime;

/*---------------------+
 | datetime            |
 +---------------------+
 | 1998-10-18 13:45:55 |
 +---------------------*/
```

```googlesql
SELECT PARSE_DATETIME('%m/%d/%Y %I:%M:%S %p', '8/30/2018 2:23:38 pm') AS datetime;

/*---------------------+
 | datetime            |
 +---------------------+
 | 2018-08-30 14:23:38 |
 +---------------------*/
```

The following example parses a `STRING` literal
containing a date in a natural language format as a
`DATETIME`.

```googlesql
SELECT PARSE_DATETIME('%A, %B %e, %Y','Wednesday, December 19, 2018')
  AS datetime;

/*---------------------+
 | datetime            |
 +---------------------+
 | 2018-12-19 00:00:00 |
 +---------------------*/
```

[datetime-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

[ISO-8601]: https://en.wikipedia.org/wiki/ISO_8601

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
