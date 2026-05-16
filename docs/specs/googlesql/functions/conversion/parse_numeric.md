---
name: PARSE_NUMERIC
dialect: googlesql
category: functions/conversion
status: implemented
source_url: docs/third_party/googlesql-docs/conversion_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/conversion_functions.md#parse_numeric
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/conversion/parse_numeric.yaml
---

# PARSE_NUMERIC

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

Verbatim copy from `docs/third_party/googlesql-docs/conversion_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PARSE_NUMERIC`

```googlesql
PARSE_NUMERIC(string_expression)
```

**Description**

Converts a `STRING` to a `NUMERIC` value.

The numeric literal contained in the string must not exceed the
[maximum precision or range][numeric-type] of the `NUMERIC` type, or an error
occurs. If the number of digits after the decimal point exceeds nine, then the
resulting `NUMERIC` value rounds
[half away from zero][half-from-zero-wikipedia] to have nine digits after the
decimal point.

```googlesql

-- This example shows how a string with a decimal point is parsed.
SELECT PARSE_NUMERIC("123.45") AS parsed;

/*--------+
 | parsed |
 +--------+
 | 123.45 |
 +--------*/

-- This example shows how a string with an exponent is parsed.
SELECT PARSE_NUMERIC("12.34E27") as parsed;

/*-------------------------------+
 | parsed                        |
 +-------------------------------+
 | 12340000000000000000000000000 |
 +-------------------------------*/

-- This example shows the rounding when digits after the decimal point exceeds 9.
SELECT PARSE_NUMERIC("1.0123456789") as parsed;

/*-------------+
 | parsed      |
 +-------------+
 | 1.012345679 |
 +-------------*/
```

This function is similar to using the [`CAST AS NUMERIC`][cast-numeric] function
except that the `PARSE_NUMERIC` function only accepts string inputs and allows
the following in the string:

+   Spaces between the sign (+/-) and the number
+   Signs (+/-) after the number

Rules for valid input strings:

<table>
  <thead>
    <tr>
      <th>Rule</th>
      <th>Example Input</th>
      <th>Output</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        The string can only contain digits, commas, decimal points and signs.
      </td>
      <td>
        "- 12,34567,89.0"
      </td>
      <td>-123456789</td>
    </tr>
    <tr>
      <td>Whitespaces are allowed anywhere except between digits.</td>
      <td>
        "  -  12.345  "
      </td>
      <td>-12.345</td>
    </tr>
    <tr>
      <td>Only digits and commas are allowed before the decimal point.</td>
      <td>
        " 12,345,678"
      </td>
      <td>12345678</td>
    </tr>
    <tr>
      <td>Only digits are allowed after the decimal point.</td>
      <td>
        "1.234 "
      </td>
      <td>1.234</td>
    </tr>
    <tr>
      <td>
        Use <code>E</code> or <code>e</code> for exponents. After
        the <code>e</code>,
        digits and a leading sign indicator are allowed.
      </td>
      <td>" 123.45e-1"</td>
      <td>12.345</td>
    </tr>
    <tr>
      <td>
        If the integer part isn't empty, then it must contain at least one
        digit.
      </td>
      <td>" 0,.12 -"</td>
      <td>-0.12</td>
    </tr>
    <tr>
      <td>
        If the string contains a decimal point, then it must contain at least
        one digit.
      </td>
      <td>" .1"</td>
      <td>0.1</td>
    </tr>
    <tr>
      <td>
        The string can't contain more than one sign.
      </td>
      <td>" 0.5 +"</td>
      <td>0.5</td>
    </tr>
  </tbody>
</table>

**Return Data Type**

`NUMERIC`

**Examples**

This example shows an input with spaces before, after, and between the
sign and the number:

```googlesql
SELECT PARSE_NUMERIC("  -  12.34 ") as parsed;

/*--------+
 | parsed |
 +--------+
 | -12.34 |
 +--------*/
```

This example shows an input with an exponent as well as the sign after the
number:

```googlesql
SELECT PARSE_NUMERIC("12.34e-1-") as parsed;

/*--------+
 | parsed |
 +--------+
 | -1.234 |
 +--------*/
```

This example shows an input with multiple commas in the integer part of the
number:

```googlesql
SELECT PARSE_NUMERIC("  1,2,,3,.45 + ") as parsed;

/*--------+
 | parsed |
 +--------+
 | 123.45 |
 +--------*/
```

This example shows an input with a decimal point and no digits in the whole
number part:

```googlesql
SELECT PARSE_NUMERIC(".1234  ") as parsed;

/*--------+
 | parsed |
 +--------+
 | 0.1234 |
 +--------*/
```

**Examples of invalid inputs**

This example is invalid because the whole number part contains no digits:

```googlesql
SELECT PARSE_NUMERIC(",,,.1234  ") as parsed;
```

This example is invalid because there are whitespaces between digits:

```googlesql
SELECT PARSE_NUMERIC("1  23.4 5  ") as parsed;
```

This example is invalid because the number is empty except for an exponent:

```googlesql
SELECT PARSE_NUMERIC("  e1 ") as parsed;
```

This example is invalid because the string contains multiple signs:

```googlesql
SELECT PARSE_NUMERIC("  - 12.3 - ") as parsed;
```

This example is invalid because the value of the number falls outside the range
of `BIGNUMERIC`:

```googlesql
SELECT PARSE_NUMERIC("12.34E100 ") as parsed;
```

This example is invalid because the string contains invalid characters:

```googlesql
SELECT PARSE_NUMERIC("$12.34") as parsed;
```

[half-from-zero-wikipedia]: https://en.wikipedia.org/wiki/Rounding#Round_half_away_from_zero

[cast-numeric]: #cast_numeric

[numeric-type]: https://github.com/google/googlesql/blob/master/docs/data-types.md#decimal_types

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/conversion_functions.md`.
