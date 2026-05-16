---
name: CAST
dialect: googlesql
category: functions/conversion
status: implemented
source_url: docs/third_party/googlesql-docs/conversion_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/conversion_functions.md#cast
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/conversion/cast.yaml
---

# CAST

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

## `CAST` 
<a id="cast"></a>

```googlesql
CAST(expression AS typename [format_clause])
```

**Description**

Cast syntax is used in a query to indicate that the result type of an
expression should be converted to some other type.

When using `CAST`, a query can fail if GoogleSQL is unable to perform
the cast. If you want to protect your queries from these types of errors, you
can use [SAFE_CAST][con-func-safecast].

Casts between supported types that don't successfully map from the original
value to the target domain produce runtime errors. For example, casting
`BYTES` to `STRING` where the byte sequence isn't valid UTF-8 results in a
runtime error.

Other examples include:

+ Casting `INT64` to `INT32` where the value overflows `INT32`.
+ Casting `STRING` to `INT32` where the `STRING` contains non-digit characters.

Some casts can include a [format clause][formatting-syntax], which provides
instructions for how to conduct the
cast. For example, you could
instruct a cast to convert a sequence of bytes to a BASE64-encoded string
instead of a UTF-8-encoded string.

The structure of the format clause is unique to each type of cast and more
information is available in the section for that cast.

**Examples**

The following query results in `"true"` if `x` is `1`, `"false"` for any other
non-`NULL` value, and `NULL` if `x` is `NULL`.

```googlesql
CAST(x=1 AS STRING)
```

### CAST AS ARRAY

```googlesql
CAST(expression AS ARRAY<element_type>)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `ARRAY`. The
`expression` parameter can represent an expression for these data types:

+ `ARRAY`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>ARRAY</code></td>
    <td><code>ARRAY</code></td>
    <td>
      
      The element types of the input
      array must be castable to the
      element types of the target array.
      For example, casting from type
      <code>ARRAY&lt;INT64&gt;</code> to
      <code>ARRAY&lt;DOUBLE&gt;</code> or
      <code>ARRAY&lt;STRING&gt;</code> is valid;
      casting from type <code>ARRAY&lt;INT64&gt;</code>
      to <code>ARRAY&lt;BYTES&gt;</code> isn't valid.
      
    </td>
  </tr>
</table>

### CAST AS BIGNUMERIC 
<a id="cast_bignumeric"></a>

```googlesql
CAST(expression AS BIGNUMERIC)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `BIGNUMERIC`. The
`expression` parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `FLOAT`
+ `DOUBLE`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td>Floating Point</td>
    <td><code>BIGNUMERIC</code></td>
    <td>
      The floating point number will round
      <a href="https://en.wikipedia.org/wiki/Rounding#Round_half_away_from_zero">half away from zero</a>.

      Casting a <code>NaN</code>, <code>+inf</code> or
      <code>-inf</code> will return an error. Casting a value outside the range
      of <code>BIGNUMERIC</code> returns an overflow error.
    </td>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>BIGNUMERIC</code></td>
    <td>
      The numeric literal contained in the string must not exceed
      the maximum precision or range of the
      <code>BIGNUMERIC</code> type, or an error will occur. If the number of
      digits after the decimal point exceeds 38, then the resulting
      <code>BIGNUMERIC</code> value will round
      <a href="https://en.wikipedia.org/wiki/Rounding#Round_half_away_from_zero">half away from zero</a>

      to have 38 digits after the decimal point.
    </td>
  </tr>
</table>

### CAST AS BOOL

```googlesql
CAST(expression AS BOOL)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `BOOL`. The
`expression` parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `BOOL`
+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td>Integer</td>
    <td><code>BOOL</code></td>
    <td>
      Returns <code>FALSE</code> if <code>x</code> is <code>0</code>,
      <code>TRUE</code> otherwise.
    </td>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>BOOL</code></td>
    <td>
      Returns <code>TRUE</code> if <code>x</code> is <code>"true"</code> and
      <code>FALSE</code> if <code>x</code> is <code>"false"</code><br />
      All other values of <code>x</code> are invalid and throw an error instead
      of casting to a boolean.<br />
      A string is case-insensitive when converting
      to a boolean.
    </td>
  </tr>
</table>

### CAST AS BYTES

```googlesql
CAST(expression AS BYTES [format_clause])
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `BYTES`. The
`expression` parameter can represent an expression for these data types:

+ `BYTES`
+ `STRING`
+ `PROTO`

**Format clause**

When an expression of one type is cast to another type, you can use the
[format clause][formatting-syntax] to provide instructions for how to conduct
the cast. You can use the format clause in this section if `expression` is a
`STRING`.

+ [Format string as bytes][format-string-as-bytes]

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>BYTES</code></td>
    <td>
      Strings are cast to bytes using UTF-8 encoding. For example,
      the string "&copy;", when cast to
      bytes, would become a 2-byte sequence with the
      hex values C2 and A9.
    </td>
  </tr>
  
  <tr>
    <td><code>PROTO</code></td>
    <td><code>BYTES</code></td>
    <td>
      Returns the proto2 wire format bytes
      of <code>x</code>.
    </td>
  </tr>
  
</table>

### CAST AS DATE

```googlesql
CAST(expression AS DATE [format_clause])
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `DATE`. The `expression`
parameter can represent an expression for these data types:

+ `STRING`
+ `DATETIME`
+ `TIMESTAMP`

**Format clause**

When an expression of one type is cast to another type, you can use the
[format clause][formatting-syntax] to provide instructions for how to conduct
the cast. You can use the format clause in this section if `expression` is a
`STRING`.

+ [Format string as date and time][format-string-as-date-time]

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>DATE</code></td>
    <td>
      When casting from string to date, the string must conform to
      the supported date literal format, and is independent of time zone. If the
      string expression is invalid or represents a date that's outside of the
      supported min/max range, then an error is produced.
    </td>
  </tr>
  
  <tr>
    <td><code>TIMESTAMP</code></td>
    <td><code>DATE</code></td>
    <td>
      Casting from a timestamp to date effectively truncates the timestamp as
      of the default time zone.
    </td>
  </tr>
  
</table>

### CAST AS DATETIME

```googlesql
CAST(expression AS DATETIME [format_clause])
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `DATETIME`. The
`expression` parameter can represent an expression for these data types:

+ `STRING`
+ `DATETIME`
+ `TIMESTAMP`

**Format clause**

When an expression of one type is cast to another type, you can use the
[format clause][formatting-syntax] to provide instructions for how to conduct
the cast. You can use the format clause in this section if `expression` is a
`STRING`.

+ [Format string as date and time][format-string-as-date-time]

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>DATETIME</code></td>
    <td>
      When casting from string to datetime, the string must conform to the
      supported datetime literal format, and is independent of time zone. If
      the string expression is invalid or represents a datetime that's outside
      of the supported min/max range, then an error is produced.
    </td>
  </tr>
  
  <tr>
    <td><code>TIMESTAMP</code></td>
    <td><code>DATETIME</code></td>
    <td>
      Casting from a timestamp to datetime effectively truncates the timestamp
      as of the default time zone.
    </td>
  </tr>
  
</table>

### CAST AS ENUM

```googlesql
CAST(expression AS ENUM)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `ENUM`. The `expression`
parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `STRING`
+ `ENUM`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>ENUM</code></td>
    <td><code>ENUM</code></td>
    <td>Must have the same enum name.</td>
  </tr>
</table>

### CAST AS Floating Point 
<a id="cast_as_floating_point"></a>

```googlesql
CAST(expression AS DOUBLE)
```

```googlesql
CAST(expression AS FLOAT)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to floating point types.
The `expression` parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `FLOAT`
+ `DOUBLE`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td>Integer</td>
    <td>Floating Point</td>
    <td>
      Returns a close but potentially not exact floating point value.
    </td>
  </tr>
  
  <tr>
    <td><code>NUMERIC</code></td>
    <td>Floating Point</td>
    <td>
      <code>NUMERIC</code> will convert to the closest floating point number
      with a possible loss of precision.
    </td>
  </tr>
  
  
  <tr>
    <td><code>BIGNUMERIC</code></td>
    <td>Floating Point</td>
    <td>
      <code>BIGNUMERIC</code> will convert to the closest floating point number
      with a possible loss of precision.
    </td>
  </tr>
  
  <tr>
    <td><code>STRING</code></td>
    <td>Floating Point</td>
    <td>
      Returns <code>x</code> as a floating point value, interpreting it as
      having the same form as a valid floating point literal.
      Also supports casts from <code>"[+,-]inf"</code> to
      <code>[,-]Infinity</code>,
      <code>"[+,-]infinity"</code> to <code>[,-]Infinity</code>, and
      <code>"[+,-]nan"</code> to <code>NaN</code>.
      Conversions are case-insensitive.
    </td>
  </tr>
</table>

### CAST AS Integer 
<a id="cast_as_integer"></a>

```googlesql
CAST(expression AS INT32)
```

```googlesql
CAST(expression AS UINT32)
```

```googlesql
CAST(expression AS INT64)
```

```googlesql
CAST(expression AS UINT64)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to integer types.
The `expression` parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `FLOAT`
+ `DOUBLE`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `ENUM`
+ `BOOL`
+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  
  <tr>
    <td>
      Floating Point
    </td>
    <td>
      Integer
    </td>
    <td>
      Returns the closest integer value.<br />
      Halfway cases such as 1.5 or -0.5 round away from zero.
    </td>
  </tr>
  <tr>
    <td><code>BOOL</code></td>
    <td>Integer</td>
    <td>
      Returns <code>1</code> if <code>x</code> is <code>TRUE</code>,
      <code>0</code> otherwise.
    </td>
  </tr>
  
  <tr>
    <td><code>STRING</code></td>
    <td>Integer</td>
    <td>
      A hex string can be cast to an integer. For example,
      <code>0x123</code> to <code>291</code> or <code>-0x123</code> to
      <code>-291</code>.
    </td>
  </tr>
  
</table>

**Examples**

If you are working with hex strings (`0x123`), you can cast those strings as
integers:

```googlesql
SELECT '0x123' as hex_value, CAST('0x123' as INT64) as hex_to_int;

/*-----------+------------+
 | hex_value | hex_to_int |
 +-----------+------------+
 | 0x123     | 291        |
 +-----------+------------*/
```

```googlesql
SELECT '-0x123' as hex_value, CAST('-0x123' as INT64) as hex_to_int;

/*-----------+------------+
 | hex_value | hex_to_int |
 +-----------+------------+
 | -0x123    | -291       |
 +-----------+------------*/
```

### CAST AS INTERVAL

```googlesql
CAST(expression AS INTERVAL)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `INTERVAL`. The
`expression` parameter can represent an expression for these data types:

+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>INTERVAL</code></td>
    <td>
      When casting from string to interval, the string must conform to either
      <a href="https://en.wikipedia.org/wiki/ISO_8601#Durations">ISO 8601 Duration</a>

      standard or to interval literal
      format 'Y-M D H:M:S.F'. Partial interval literal formats are also accepted
      when they aren't ambiguous, for example 'H:M:S'.
      If the string expression is invalid or represents an interval that is
      outside of the supported min/max range, then an error is produced.
    </td>
  </tr>
</table>

**Examples**

```googlesql
SELECT input, CAST(input AS INTERVAL) AS output
FROM UNNEST([
  '1-2 3 10:20:30.456',
  '1-2',
  '10:20:30',
  'P1Y2M3D',
  'PT10H20M30,456S'
]) input

/*--------------------+--------------------+
 | input              | output             |
 +--------------------+--------------------+
 | 1-2 3 10:20:30.456 | 1-2 3 10:20:30.456 |
 | 1-2                | 1-2 0 0:0:0        |
 | 10:20:30           | 0-0 0 10:20:30     |
 | P1Y2M3D            | 1-2 3 0:0:0        |
 | PT10H20M30,456S    | 0-0 0 10:20:30.456 |
 +--------------------+--------------------*/
```

### CAST AS NUMERIC 
<a id="cast_numeric"></a>

```googlesql
CAST(expression AS NUMERIC)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `NUMERIC`. The
`expression` parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `FLOAT`
+ `DOUBLE`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>Floating Point</code></td>
    <td><code>NUMERIC</code></td>
    <td>
      The floating point number will round
      <a href="https://en.wikipedia.org/wiki/Rounding#Round_half_away_from_zero">half away from zero</a>.

      Casting a <code>NaN</code>, <code>+inf</code> or
      <code>-inf</code> will return an error. Casting a value outside the range
      of <code>NUMERIC</code> returns an overflow error.
    </td>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>NUMERIC</code></td>
    <td>
      The numeric literal contained in the string must not exceed
      the maximum precision or range of the <code>NUMERIC</code>
      type, or an error will occur. If the number of digits
      after the decimal point exceeds nine, then the resulting
      <code>NUMERIC</code> value will round
      <a href="https://en.wikipedia.org/wiki/Rounding#Round_half_away_from_zero">half away from zero</a>.

      to have nine digits after the decimal point.
    </td>
  </tr>
</table>

### CAST AS PROTO

```googlesql
CAST(expression AS PROTO)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `PROTO`. The
`expression` parameter can represent an expression for these data types:

+ `STRING`
+ `BYTES`
+ `PROTO`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>PROTO</code></td>
    <td>
      Returns the protocol buffer that results from parsing
      from proto2 text format.<br />
      Throws an error if parsing fails, e.g., if not all required fields are
      set.
    </td>
  </tr>
  <tr>
    <td><code>BYTES</code></td>
    <td><code>PROTO</code></td>
    <td>
      Returns the protocol buffer that results from parsing
      <code>x</code> from the proto2 wire format.<br />
      Throws an error if parsing fails, e.g., if not all required fields are
      set.
    </td>
  </tr>
  <tr>
    <td><code>PROTO</code></td>
    <td><code>PROTO</code></td>
    <td>Must have the same protocol buffer name.</td>
  </tr>
</table>

**Example**

This example references a protocol buffer called `Award`.

```proto
message Award {
  required int32 year = 1;
  optional int32 month = 2;
  repeated Type type = 3;

  message Type {
    optional string award_name = 1;
    optional string category = 2;
  }
}
```

```googlesql
SELECT
  CAST(
    '''
    year: 2001
    month: 9
    type { award_name: 'Best Artist' category: 'Artist' }
    type { award_name: 'Best Album' category: 'Album' }
    '''
    AS googlesql.examples.music.Award)
  AS award_col

/*---------------------------------------------------------+
 | award_col                                               |
 +---------------------------------------------------------+
 | {                                                       |
 |   year: 2001                                            |
 |   month: 9                                              |
 |   type { award_name: "Best Artist" category: "Artist" } |
 |   type { award_name: "Best Album" category: "Album" }   |
 | }                                                       |
 +---------------------------------------------------------*/
```

### CAST AS RANGE

```googlesql
CAST(expression AS RANGE)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `RANGE`. The
`expression` parameter can represent an expression for these data types:

+ `STRING`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>RANGE</code></td>
    <td>
      When casting from string to range, the string must conform to the
      supported range literal format. If the string expression is invalid or
      represents a range that's outside of the supported subtype min/max range,
      then an error is produced.
    </td>
  </tr>
</table>

**Examples**

```googlesql
SELECT CAST(
  '[2020-01-01, 2020-01-02)'
  AS RANGE<DATE>) AS string_to_range

/*----------------------------------------+
 | string_to_range                        |
 +----------------------------------------+
 | [DATE '2020-01-01', DATE '2020-01-02') |
 +----------------------------------------*/
```

```googlesql
SELECT CAST(
  '[2014-09-27 12:30:00.45, 2016-10-17 11:15:00.33)'
  AS RANGE<DATETIME>) AS string_to_range

/*------------------------------------------------------------------------+
 | string_to_range                                                        |
 +------------------------------------------------------------------------+
 | [DATETIME '2014-09-27 12:30:00.45', DATETIME '2016-10-17 11:15:00.33') |
 +------------------------------------------------------------------------*/
```

```googlesql
SELECT CAST(
  '[2014-09-27 12:30:00+08, 2016-10-17 11:15:00+08)'
  AS RANGE<TIMESTAMP>) AS string_to_range

-- Results depend upon where this query was executed.
/*--------------------------------------------------------------------------+
 | string_to_range                                                          |
 +--------------------------------------------------------------------------+
 | [TIMESTAMP '2014-09-27 12:30:00+08', TIMESTAMP '2016-10-17 11:15:00+08') |
 +--------------------------------------------------------------------------*/
```

```googlesql
SELECT CAST(
  '[UNBOUNDED, 2020-01-02)'
  AS RANGE<DATE>) AS string_to_range

/*--------------------------------+
 | string_to_range                |
 +--------------------------------+
 | [UNBOUNDED, DATE '2020-01-02') |
 +--------------------------------*/
```

```googlesql
SELECT CAST(
  '[2020-01-01, NULL)'
  AS RANGE<DATE>) AS string_to_range

/*--------------------------------+
 | string_to_range                |
 +--------------------------------+
 | [DATE '2020-01-01', UNBOUNDED) |
 +--------------------------------*/
```

### CAST AS STRING 
<a id="cast_as_string"></a>

```googlesql
CAST(expression AS STRING [format_clause [AT TIME ZONE timezone_expr]])
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `STRING`. The
`expression` parameter can represent an expression for these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `FLOAT`
+ `DOUBLE`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `ENUM`
+ `BOOL`
+ `BYTES`
+ `PROTO`
+ `TIME`
+ `DATE`
+ `DATETIME`
+ `TIMESTAMP`
+ `RANGE`
+ `INTERVAL`
+ `STRING`

**Format clause**

When an expression of one type is cast to another type, you can use the
[format clause][formatting-syntax] to provide instructions for how to conduct
the cast. You can use the format clause in this section if `expression` is one
of these data types:

+ `INT32`
+ `UINT32`
+ `INT64`
+ `UINT64`
+ `FLOAT`
+ `DOUBLE`
+ `NUMERIC`
+ `BIGNUMERIC`
+ `BYTES`
+ `TIME`
+ `DATE`
+ `DATETIME`
+ `TIMESTAMP`

The format clause for `STRING` has an additional optional clause called
`AT TIME ZONE timezone_expr`, which you can use to specify a specific time zone
to use during formatting of a `TIMESTAMP`. If this optional clause isn't
included when formatting a `TIMESTAMP`, the default time zone,
which is implementation defined, is used.

For more information, see the following topics:

+ [Format bytes as string][format-bytes-as-string]
+ [Format date and time as string][format-date-time-as-string]
+ [Format numeric type as string][format-numeric-type-as-string]

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td>Floating Point</td>
    <td><code>STRING</code></td>
    <td>Returns an approximate string representation. A returned
    <code>NaN</code> or <code>0</code> will not be signed.<br />
    </td>
  </tr>
  <tr>
    <td><code>BOOL</code></td>
    <td><code>STRING</code></td>
    <td>
      Returns <code>"true"</code> if <code>x</code> is <code>TRUE</code>,
      <code>"false"</code> otherwise.</td>
  </tr>
  <tr>
    <td><code>BYTES</code></td>
    <td><code>STRING</code></td>
    <td>
      Returns <code>x</code> interpreted as a UTF-8 string.<br />
      For example, the bytes literal
      <code>b'\xc2\xa9'</code>, when cast to a string,
      is interpreted as UTF-8 and becomes the unicode character "&copy;".<br />
      An error occurs if <code>x</code> isn't valid UTF-8.</td>
  </tr>
  
  <tr>
    <td><code>ENUM</code></td>
    <td><code>STRING</code></td>
    <td>
      Returns the canonical enum value name of
      <code>x</code>.<br />
      If an enum value has multiple names (aliases),
      the canonical name/alias for that value is used.</td>
  </tr>
  
  
  <tr>
    <td><code>PROTO</code></td>
    <td><code>STRING</code></td>
    <td>Returns the proto2 text format representation of <code>x</code>.</td>
  </tr>
  
  
  <tr>
    <td><code>TIME</code></td>
    <td><code>STRING</code></td>
    <td>
      Casting from a time type to a string is independent of time zone and
      is of the form <code>HH:MM:SS</code>.
    </td>
  </tr>
  
  
  <tr>
    <td><code>DATE</code></td>
    <td><code>STRING</code></td>
    <td>
      Casting from a date type to a string is independent of time zone and is
      of the form <code>YYYY-MM-DD</code>.
    </td>
  </tr>
  
  
  <tr>
    <td><code>DATETIME</code></td>
    <td><code>STRING</code></td>
    <td>
      Casting from a datetime type to a string is independent of time zone and
      is of the form <code>YYYY-MM-DD HH:MM:SS</code>.
    </td>
  </tr>
  
  
  <tr>
    <td><code>TIMESTAMP</code></td>
    <td><code>STRING</code></td>
    <td>
      When casting from timestamp types to string, the timestamp is interpreted
      using the default time zone, which is implementation defined. The number of
      subsecond digits produced depends on the number of trailing zeroes in the
      subsecond part: the CAST function will truncate zero, three, or six
      digits.
    </td>
  </tr>
  
  <tr>
    <td><code>INTERVAL</code></td>
    <td><code>STRING</code></td>
    <td>
      Casting from an interval to a string is of the form
      <code>Y-M D H:M:S</code>.
    </td>
  </tr>
  
</table>

**Examples**

```googlesql
SELECT CAST(CURRENT_DATE() AS STRING) AS current_date

/*---------------+
 | current_date  |
 +---------------+
 | 2021-03-09    |
 +---------------*/
```

```googlesql
SELECT CAST(CURRENT_DATE() AS STRING FORMAT 'DAY') AS current_day

/*-------------+
 | current_day |
 +-------------+
 | MONDAY      |
 +-------------*/
```

```googlesql
SELECT CAST(
  TIMESTAMP '2008-12-25 00:00:00+00:00'
  AS STRING FORMAT 'YYYY-MM-DD HH24:MI:SS TZH:TZM') AS date_time_to_string

-- Results depend upon where this query was executed.
/*------------------------------+
 | date_time_to_string          |
 +------------------------------+
 | 2008-12-24 16:00:00 -08:00   |
 +------------------------------*/
```

```googlesql
SELECT CAST(
  TIMESTAMP '2008-12-25 00:00:00+00:00'
  AS STRING FORMAT 'YYYY-MM-DD HH24:MI:SS TZH:TZM'
  AT TIME ZONE 'Asia/Kolkata') AS date_time_to_string

-- Because the time zone is specified, the result is always the same.
/*------------------------------+
 | date_time_to_string          |
 +------------------------------+
 | 2008-12-25 05:30:00 +05:30   |
 +------------------------------*/
```

```googlesql
SELECT CAST(INTERVAL 3 DAY AS STRING) AS interval_to_string

/*--------------------+
 | interval_to_string |
 +--------------------+
 | 0-0 3 0:0:0        |
 +--------------------*/
```

```googlesql
SELECT CAST(
  INTERVAL "1-2 3 4:5:6.789" YEAR TO SECOND
  AS STRING) AS interval_to_string

/*--------------------+
 | interval_to_string |
 +--------------------+
 | 1-2 3 4:5:6.789    |
 +--------------------*/
```

### CAST AS STRUCT

```googlesql
CAST(expression AS STRUCT)
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `STRUCT`. The
`expression` parameter can represent an expression for these data types:

+ `STRUCT`

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRUCT</code></td>
    <td><code>STRUCT</code></td>
    <td>
      Allowed if the following conditions are met:<br />
      <ol>
        <li>
          The two structs have the same number of
          fields.
        </li>
        <li>
          The original struct field types can be
          explicitly cast to the corresponding target
          struct field types (as defined by field
          order, not field name).
        </li>
      </ol>
    </td>
  </tr>
</table>

### CAST AS TIME

```googlesql
CAST(expression AS TIME [format_clause])
```

**Description**

GoogleSQL supports [casting][con-func-cast] to TIME. The `expression`
parameter can represent an expression for these data types:

+ `STRING`
+ `TIME`
+ `DATETIME`
+ `TIMESTAMP`

**Format clause**

When an expression of one type is cast to another type, you can use the
[format clause][formatting-syntax] to provide instructions for how to conduct
the cast. You can use the format clause in this section if `expression` is a
`STRING`.

+ [Format string as date and time][format-string-as-date-time]

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>TIME</code></td>
    <td>
      When casting from string to time, the string must conform to
      the supported time literal format, and is independent of time zone. If the
      string expression is invalid or represents a time that's outside of the
      supported min/max range, then an error is produced.
    </td>
  </tr>
</table>

### CAST AS TIMESTAMP

```googlesql
CAST(expression AS TIMESTAMP [format_clause [AT TIME ZONE timezone_expr]])
```

**Description**

GoogleSQL supports [casting][con-func-cast] to `TIMESTAMP`. The
`expression` parameter can represent an expression for these data types:

+ `STRING`
+ `DATETIME`
+ `TIMESTAMP`

**Format clause**

When an expression of one type is cast to another type, you can use the
[format clause][formatting-syntax] to provide instructions for how to conduct
the cast. You can use the format clause in this section if `expression` is a
`STRING`.

+ [Format string as date and time][format-string-as-date-time]

The format clause for `TIMESTAMP` has an additional optional clause called
`AT TIME ZONE timezone_expr`, which you can use to specify a specific time zone
to use during formatting. If this optional clause isn't included, the default
time zone, which is implementation defined, is used.

**Conversion rules**

<table>
  <tr>
    <th>From</th>
    <th>To</th>
    <th>Rule(s) when casting <code>x</code></th>
  </tr>
  <tr>
    <td><code>STRING</code></td>
    <td><code>TIMESTAMP</code></td>
    <td>
      When casting from string to a timestamp, <code>string_expression</code>
      must conform to the supported timestamp literal formats, or else a runtime
      error occurs. The <code>string_expression</code> may itself contain a
      time zone.
      <br /><br />
      If there is a time zone in the <code>string_expression</code>, that
      time zone is used for conversion, otherwise the default time zone,
      which is implementation defined, is used. If the string has fewer than six digits,
      then it's implicitly widened.
      <br /><br />
      An error is produced if the <code>string_expression</code> is invalid,
      has more than six subsecond digits (i.e., precision greater than
      microseconds), or represents a time outside of the supported timestamp
      range.
    </td>
  </tr>
  
  <tr>
    <td><code>DATE</code></td>
    <td><code>TIMESTAMP</code></td>
    <td>
      Casting from a date to a timestamp interprets <code>date_expression</code>
      as of midnight (start of the day) in the default time zone,
      which is implementation defined.
    </td>
  </tr>
  
  
  <tr>
    <td><code>DATETIME</code></td>
    <td><code>TIMESTAMP</code></td>
    <td>
      Casting from a datetime to a timestamp interprets
      <code>datetime_expression</code> in the default time zone,
      which is implementation defined.
      <br /><br />
      Most valid datetime values have exactly one corresponding timestamp
      in each time zone. However, there are certain combinations of valid
      datetime values and time zones that have zero or two corresponding
      timestamp values. This happens in a time zone when clocks are set forward
      or set back, such as for Daylight Savings Time.
      When there are two valid timestamps, the earlier one is used.
      When there is no valid timestamp, the length of the gap in time
      (typically one hour) is added to the datetime.
    </td>
  </tr>
  
</table>

**Examples**

The following example casts a string-formatted timestamp as a timestamp:

```googlesql
SELECT CAST("2020-06-02 17:00:53.110+00:00" AS TIMESTAMP) AS as_timestamp

-- Results depend upon where this query was executed.
/*----------------------------+
 | as_timestamp               |
 +----------------------------+
 | 2020-06-03 00:00:53.110+00 |
 +----------------------------*/
```

The following examples cast a string-formatted date and time as a timestamp.
These examples return the same output as the previous example.

```googlesql
SELECT CAST('06/02/2020 17:00:53.110' AS TIMESTAMP FORMAT 'MM/DD/YYYY HH24:MI:SS.FF3' AT TIME ZONE 'UTC') AS as_timestamp
```

```googlesql
SELECT CAST('06/02/2020 17:00:53.110' AS TIMESTAMP FORMAT 'MM/DD/YYYY HH24:MI:SS.FF3' AT TIME ZONE '+00') AS as_timestamp
```

```googlesql
SELECT CAST('06/02/2020 17:00:53.110 +00' AS TIMESTAMP FORMAT 'MM/DD/YYYY HH24:MI:SS.FF3 TZH') AS as_timestamp
```

[formatting-syntax]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#formatting_syntax

[format-string-as-bytes]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_string_as_bytes

[format-bytes-as-string]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_bytes_as_string

[format-date-time-as-string]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_date_time_as_string

[format-string-as-date-time]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_string_as_datetime

[format-numeric-type-as-string]: https://github.com/google/googlesql/blob/master/docs/format-elements.md#format_numeric_type_as_string

[con-func-cast]: #cast

[con-func-safecast]: #safe_casting

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/conversion_functions.md`.
