---
name: FORMAT_DATE
dialect: googlesql
category: functions/date
status: implemented
source_url: docs/third_party/googlesql-docs/date_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/date_functions.md#format_date
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/date/format_date.yaml
---

# FORMAT_DATE

## Summary

Formats a `DATE` value as a `STRING` according to a specified format string.

## Signatures

- `FORMAT_DATE(format_string, date_expr)`

## Behavior

- Returns a `STRING`.
- `format_string` is a `STRING` whose contents are interpreted as date/time format elements applied to `date_expr`.
- `date_expr` must be a `DATE` value representing the date to format.
- Format elements are the date/time format elements documented at the upstream `format_elements_date_time` reference.

## Examples

```googlesql
SELECT FORMAT_DATE('%x', DATE '2008-12-25') AS US_format;
-- expected: 12/25/08
```

```googlesql
SELECT FORMAT_DATE('%b-%d-%Y', DATE '2008-12-25') AS formatted;
-- expected: Dec-25-2008
```

```googlesql
SELECT FORMAT_DATE('%b %Y', DATE '2008-12-25') AS formatted;
-- expected: Dec 2008
```

## Edge cases

- Upstream does not enumerate explicit NULL or error behaviour for this function; behaviour outside of a valid `STRING` format and a valid `DATE` argument is not specified by the reference block.
- Only format elements listed in the upstream date/time format-elements reference are guaranteed to be honoured.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/date_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `FORMAT_DATE`

```googlesql
FORMAT_DATE(format_string, date_expr)
```

**Description**

Formats a `DATE` value according to a specified format string.

**Definitions**

+   `format_string`: A `STRING` value that contains the
    [format elements][date-format-elements] to use with `date_expr`.
+   `date_expr`: A `DATE` value that represents the date to format.

**Return Data Type**

`STRING`

**Examples**

```googlesql
SELECT FORMAT_DATE('%x', DATE '2008-12-25') AS US_format;

/*------------+
 | US_format  |
 +------------+
 | 12/25/08   |
 +------------*/
```

```googlesql
SELECT FORMAT_DATE('%b-%d-%Y', DATE '2008-12-25') AS formatted;

/*-------------+
 | formatted   |
 +-------------+
 | Dec-25-2008 |
 +-------------*/
```

```googlesql
SELECT FORMAT_DATE('%b %Y', DATE '2008-12-25') AS formatted;

/*-------------+
 | formatted   |
 +-------------+
 | Dec 2008    |
 +-------------*/
```

[date-format-elements]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_elements_date_time

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/date_functions.md`.
