---
name: PROTOCOL BUFFER
dialect: googlesql
category: types
status: partial
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#protocol-buffer-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/protocol_buffer.yaml
---

# PROTOCOL BUFFER

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

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Protocol buffer type 
<a id="protocol_buffer_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>PROTO</code></td>
<td>An instance of protocol buffer.</td>
</tr>
</tbody>
</table>

Protocol buffers provide structured data types with a defined serialization
format and cross-language support libraries. Protocol buffer message types can
contain optional, required, or repeated fields, including nested messages. For
more information, see the [Protocol Buffers Developer Guide][protocol-buffers-dev-guide].

Protocol buffer message types behave similarly to [struct types](#struct_type),
and support similar operations like reading field values by name. Protocol
buffer types are always named types, and can be referred to by their
fully-qualified protocol buffer name (i.e. `package.ProtoName`). Protocol
buffers support some additional behavior beyond structs, like default field
values, defining a column type, and checking for the presence of
optional fields.

Protocol buffer [enum types](#enum_type) are also available and can be
referenced using the fully-qualified enum type name.

To learn more about using protocol buffers in GoogleSQL, see
[Work with protocol buffers][protocol-buffers].

### Constructing a protocol buffer 
<a id="constructing_a_proto"></a>

You can construct a protocol buffer using the [`NEW`][new-operator] operator or
the [`SELECT AS typename`][select-as-typename] statement. Regardless of the
method that you choose, the resulting protocol buffer is the same.

#### `NEW protocol_buffer {...}` {: #using_new_map_constructor }

You can create a protocol buffer using the [`NEW`][new-operator]
operator with a map constructor:

```googlesql
NEW protocol_buffer {
  field_name: literal_or_expression
  field_name { ... }
  repeated_field_name: [literal_or_expression, ... ]
  map_field_name: [{key: literal_or_expression value: literal_or_expression}, ...],
  (extension_name): literal_or_expression
}
```

Where:

+ `protocol_buffer`: The full protocol buffer name including the package name.
+ `field_name`: The name of a field.
+ `literal_or_expression`: The field value.
+  `map_field_name`: The name of a map-typed field. The value is a list of
   key/value pair entries for the map.
+  `extension_name`: The name of the proto extension, including the package
   name.

**Example**

```googlesql
NEW googlesql.examples.astronomy.Planet {
  planet_name: 'Jupiter'
  facts: {
    length_of_day: 9.93
    distance_to_sun: 5.2 * ASTRONOMICAL_UNIT
    has_rings: TRUE
  }
  major_moons: [
    { moon_name: 'Io' },
    { moon_name: 'Europa' },
    { moon_name: 'Ganymede' },
    { moon_name: 'Callisto'}
  ]
  minor_moons: (
    SELECT ARRAY_AGG(moon_name)
    FROM SolarSystemMoons
    WHERE
      planet_name = 'Jupiter'
      AND circumference < 3121
  )
  count_of_space_probe_photos: (
    GALILEO_PHOTOS
    + JUNO_PHOTOS
    + NEW_HORIZONS_PHOTOS
    + CASSINI_PHOTOS
    + ULYSSES_PHOTOS
    + VOYAGER_1_PHOTOS
    + VOYAGER_2_PHOTOS
    + PIONEER_10_PHOTOS
    + PIONEER_11_PHOTOS
  ),
  (UniverseExtraInfo.extension) {
    ...
  }
}
```

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

> NOTE: The syntax is very similar to the Protocol Buffer Text Format
>  syntax.
> The differences are:
>
> +   Values can be arbitrary SQL expressions instead of having to be literals.
> +   Repeated fields are written as `x_array: [1, 2, 3]` instead of `x_array:`
>     appearing multiple times.
> +   Extensions use parentheses instead of square brackets.

<!-- mdlint on -->

When using this syntax, the following rules apply:

+   The field values must be expressions that are implicitly coercible or
    literal-coercible to the type of the corresponding protocol buffer field.
+   Commas between fields are optional.
+   Extension names must have parentheses around the path and must have a comma
    preceding the extension field (unless it's the first field).
+   A colon is required between field name and values unless the value is a map
    constructor.
+   The `NEW protocol_buffer` prefix is optional if the protocol buffer type can
    be inferred from the context.
+   The type of submessages inside the map constructor can be inferred.

**Examples**

Simple:

```googlesql
SELECT
  key,
  name,
  NEW googlesql.examples.music.Chart { rank: 1 chart_name: '2' }
```

Nested messages and arrays:

```googlesql
SELECT
  NEW googlesql.examples.music.Album {
    album_name: 'New Moon'
    singer {
      nationality: 'Canadian'
      residence: [ { city: 'Victoria' }, { city: 'Toronto' } ]
    }
    song: ['Sandstorm', 'Wait']
  }
```

With an extension field (note a comma is required before the extension field):

```googlesql
SELECT
  NEW googlesql.examples.music.Album {
    album_name: 'New Moon',
    (googlesql.examples.music.downloads): 30
  }
```

Non-literal expressions as values:

```googlesql
SELECT
  NEW googlesql.examples.music.Chart {
    rank: (SELECT COUNT(*) FROM TableName WHERE foo = 'bar')
    chart_name: CONCAT('best', 'hits')
  }
```

The following examples infers the protocol buffer data type from context:

+   From `ARRAY` constructor:

    ```googlesql
    SELECT
      ARRAY<googlesql.examples.music.Chart>[
        { rank: 1 chart_name: '2' },
        { rank: 2 chart_name: '3' }]
    ```
+   From `STRUCT` constructor:

    ```googlesql
    SELECT
      STRUCT<STRING, googlesql.examples.music.Chart, INT64>(
        'foo', { rank: 1 chart_name: '2' }, 7)[1]
    ```
+   From column names through `SET`:

    +   Simple column:

    ```googlesql
    UPDATE TableName SET proto_column = { rank: 1 chart_name: '2' }
    ```

    +   Array column:

    ```googlesql
    UPDATE TableName
    SET proto_array_column = [
      { rank: 1 chart_name: '2' }, { rank: 2 chart_name: '3' }]
    ```

    +   Struct column:

    ```googlesql
    UPDATE TableName
    SET proto_struct_column = ('foo', { rank: 1 chart_name: '2' }, 7)
    ```
+   From generated column names in `CREATE`:

    ```googlesql
    CREATE TABLE TableName (
      proto_column googlesql.examples.music.Chart GENERATED ALWAYS AS (
        { rank: 1 chart_name: '2' }))
    ```
+   From column names in default values in `CREATE`:

    ```googlesql
    CREATE TABLE TableName(
      proto_column googlesql.examples.music.Chart DEFAULT (
        { rank: 1 chart_name: '2' }))
    ```
+   From return types in SQL function body:

    ```googlesql
    CREATE FUNCTION MyFunc()
    RETURNS googlesql.examples.music.Chart
    AS (
      { rank: 1 chart_name: '2' }
    )
    ```

#### `NEW protocol_buffer (...)` 
<a id="using_new"></a>

You can create a protocol buffer using the [`NEW`][new-operator] operator with a
parenthesized list of arguments and aliases to specify field names:

```googlesql
NEW protocol_buffer(field [AS alias], ...)
```

**Example**

```googlesql
SELECT
  key,
  name,
  NEW googlesql.examples.music.Chart(key AS rank, name AS chart_name)
FROM
  (SELECT 1 AS key, "2" AS name);
```

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

When using this syntax, the following rules apply:

+   All field expressions must have an [explicit alias][explicit-alias] or end
    with an identifier. For example, the expression `a.b.c` has the [implicit
    alias][implicit-alias] `c`.
+   `NEW` matches fields by alias to the field names of the protocol buffer.
    Aliases must be unique.
+   The expressions must be implicitly coercible or literal-coercible to the
    type of the corresponding protocol buffer field.

To create a protocol buffer with an extension, use this syntax:

```googlesql
NEW protocol_buffer(expression AS (path.to.extension), ...)
```

+   For `path.to.extension`, provide the path to the extension. Place the
    extension path inside parentheses.
+   `expression` provides the value to set for the extension. `expression` must
    be of the same type as the extension or [coercible to that
    type][conversion-rules].

    Example:

    ```googlesql
    SELECT
     NEW googlesql.examples.music.Album (
       album AS album_name,
       count AS (googlesql.examples.music.downloads)
     )
     FROM (SELECT 'New Moon' AS album, 30 AS count);

    /*---------------------------------------------------+
     | $col1                                             |
     +---------------------------------------------------+
     | {album_name: 'New Moon' [...music.downloads]: 30} |
     +---------------------------------------------------*/
    ```
+   If `path.to.extension` points to a nested protocol buffer extension, `expr1`
    provides an instance or a text format string of that protocol buffer.

    Example:

    ```googlesql
    SELECT
     NEW googlesql.examples.music.Album(
       'New Moon' AS album_name,
       NEW googlesql.examples.music.AlbumExtension(
        DATE(1956,1,1) AS release_date
       )
     AS (googlesql.examples.music.AlbumExtension.album_extension));

    /*---------------------------------------------+
     | $col1                                       |
     +---------------------------------------------+
     | album_name: "New Moon"                      |
     | [...music.AlbumExtension.album_extension] { |
     |   release_date: -5114                       |
     | }                                           |
     +---------------------------------------------*/
    ```

#### `SELECT AS typename` 
<a id="select_as_typename_proto"></a>

The [`SELECT AS typename`][select-as-typename] statement can produce a
value table where the row type is a specific named protocol buffer type.

`SELECT AS` doesn't support setting protocol buffer extensions. To do so, use
the [`NEW`][new-keyword] keyword instead. For example,  to create a
protocol buffer with an extension, change a query like this:

```googlesql
SELECT AS typename field1, field2, ...
```

to a query like this:

```googlesql
SELECT AS VALUE NEW ProtoType(field1, field2, field3 AS (path.to.extension), ...)
```

### Limited comparisons for protocol buffer values

Direct comparison of protocol buffers isn't supported. There are a few
alternative solutions:

+ One way to compare protocol buffers is to do a pair-wise
  comparison between the fields of the protocol buffers. This can also be used
  to `GROUP BY` or `ORDER BY` protocol buffer fields.
+ To get a simple approximation comparison, cast protocol buffer to
  string. This applies lexicographical ordering for numeric fields.

[protocol-buffers-dev-guide]: https://developers.google.com/protocol-buffers/docs/overview

[protocol-buffers]: https://github.com/google/googlesql/blob/master/docs/protocol-buffers.md

[new-operator]: https://github.com/google/googlesql/blob/master/docs/operators.md#new_operator

[select-as-typename]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#select_as_typename

[new-keyword]: #using_new

[explicit-alias]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#explicit_alias_syntax

[implicit-alias]: https://github.com/google/googlesql/blob/master/docs/query-syntax.md#implicit_aliases

[conversion-rules]: https://github.com/google/googlesql/blob/master/docs/conversion_rules.md

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
