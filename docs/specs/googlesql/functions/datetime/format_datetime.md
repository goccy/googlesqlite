---
name: FORMAT_DATETIME
dialect: googlesql
category: functions/datetime
status: implemented
source_url: docs/third_party/googlesql-docs/datetime_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/datetime_functions.md#format_datetime
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/datetime/format_datetime.yaml
---

# FORMAT_DATETIME

## Summary

Formats a `DATETIME` value according to a specified format string and returns
the result as a `STRING`.

## Signatures

- `FORMAT_DATETIME(format_string, datetime_expr)`

## Behavior

- Returns a `STRING`.
- `format_string` is a `STRING` containing the supported datetime format
  elements that drive the rendering of `datetime_expr`.
- `datetime_expr` is the `DATETIME` value representing the date and time to
  format.
- Format elements follow the googlesql datetime format-elements reference.

## Examples

```googlesql
SELECT FORMAT_DATETIME("%c", DATETIME "2008-12-25 15:30:00") AS formatted;
-- expected: Thu Dec 25 15:30:00 2008
```

```googlesql
SELECT FORMAT_DATETIME("%b-%d-%Y", DATETIME "2008-12-25 15:30:00") AS formatted;
-- expected: Dec-25-2008
```

```googlesql
SELECT FORMAT_DATETIME("%b %Y", DATETIME "2008-12-25 15:30:00") AS formatted;
-- expected: Dec 2008
```

## Edge cases

- The upstream reference does not enumerate explicit NULL or error behaviour
  for this function; consult the datetime format-elements documentation for
  rules governing unsupported or malformed format specifiers.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/datetime_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `FORMAT_DATETIME`

```googlesql
FORMAT_DATETIME(format_string, datetime_expr)
```

**Description**

Formats a `DATETIME` value according to a specified format string.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][datetime-format-elements] to use with
    `datetime_expr`.
+   `datetime_expr`: A `DATETIME` value that represents the date and time to
    format.

**Return Data Type**

`STRING`

**Examples**

```googlesql
SELECT
  FORMAT_DATETIME("%c", DATETIME "2008-12-25 15:30:00")
  AS formatted;

/*--------------------------+
 | formatted                |
 +--------------------------+
 | Thu Dec 25 15:30:00 2008 |
 +--------------------------*/
```

```googlesql
SELECT
  FORMAT_DATETIME("%b-%d-%Y", DATETIME "2008-12-25 15:30:00")
  AS formatted;

/*-------------+
 | formatted   |
 +-------------+
 | Dec-25-2008 |
 +-------------*/
```

```googlesql
SELECT
  FORMAT_DATETIME("%b %Y", DATETIME "2008-12-25 15:30:00")
  AS formatted;

/*-------------+
 | formatted   |
 +-------------+
 | Dec 2008    |
 +-------------*/
```

[datetime-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/datetime_functions.md`.
