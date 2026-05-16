---
name: STRUCT
dialect: googlesql
category: types
status: implemented
notes: |
  Testdata sourced verbatim from googlesql-wasm `compliance/testdata/struct_queries.test` and `compliance/testdata/struct_positional_accessor.test`, translated to our brace short-form (`{v1, v2, ...}` — values only; field names live on the column type).
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#struct-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/struct.yaml
---

# STRUCT

## Summary

`STRUCT` is a container of ordered fields, each with a required type and an optional field name. Field types may be arbitrarily complex, including nested structs and arrays.

## Signatures

- `STRUCT<T>` — type declaration, where `T` is a comma-separated list of `[field_name] field_type` entries (e.g. `STRUCT<INT64>`, `STRUCT<a INT64, b STRING>`).
- `(expr1, expr2 [, ... ])` — tuple syntax constructor; produces an anonymous struct with anonymous fields. Requires at least two expressions.
- `STRUCT( expr1 [AS field_name] [, ... ])` — typeless constructor; field types are inferred from expressions and `AS` may name fields.
- `STRUCT<[field_name] field_type, ...>( expr1 [, ... ])` — typed constructor; the output type is exactly the declared struct type and inputs are coerced to the declared field types.

## Behavior

- Fields are ordered; each field has a required type and an optional name.
- Struct types are declared with angle brackets `<` and `>` and may nest arbitrarily, including other structs and arrays.
- In tuple syntax, when bare column references are used, the struct field name is derived from the column name and the field type from the column's data type.
- In typeless `STRUCT(...)` syntax, duplicate field names are allowed; unnamed fields are anonymous and cannot be referenced by name.
- In typed `STRUCT<...>(...)` syntax, the number of input expressions must match the declared field count, inputs are (literal-)coerced to the declared field types, and `AS alias` is not allowed on the inputs.
- Struct values support `=`, `!=`/`<>`, and `[NOT] IN`; these compare fields pairwise in ordinal order, ignoring field names.

## Examples

```sql
SELECT STRUCT(1 AS a, 'abc' AS b);
-- output type: STRUCT<a INT64, b STRING>
```

```sql
SELECT STRUCT<x INT64, y STRING>(1, 'abc');
-- output type: STRUCT<x INT64, y STRING>
```

```sql
SELECT *
FROM T
WHERE (Key1, Key2) IN ((12, 34), (56, 78));
-- tuple-syntax struct comparison on multi-part keys
```

## Edge cases

- Tuple syntax `(expr)` with a single expression is indistinguishable from a parenthesised expression; at least two expressions are required to form a struct.
- `STRUCT()` produces an empty struct of type `STRUCT<>`.
- Struct values can be `NULL`, and individual field values can also be `NULL`.
- Typed struct syntax raises an error if input types are not coercible to the declared field types, or if `AS alias` appears on an input expression (e.g. `STRUCT<x INT64>(5 AS x)`).
- Equality and `IN` comparisons ignore field names; to compare by name, compare individual fields directly.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Struct type 
<a id="struct_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>STRUCT</code></td>
<td>Container of ordered fields each with a type (required) and field name
(optional).</td>
</tr>
</tbody>
</table>

To learn more about the literal representation of a struct type,
see [Struct literals][struct-literals].

### Declaring a struct type

```
STRUCT<T>
```

Struct types are declared using the angle brackets (`<` and `>`). The type of
the elements of a struct can be arbitrarily complex.

**Examples**

<table>
<thead>
<tr>
<th>Type Declaration</th>
<th>Meaning</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>
STRUCT&lt;INT64&gt;
</code>
</td>
<td>Simple struct with a single unnamed 64-bit integer field.</td>
</tr>
<tr>
<td style="white-space:nowrap">
<code>
STRUCT&lt;x STRUCT&lt;y INT64, z INT64&gt;&gt;
</code>
</td>
<td>A struct with a nested struct named <code>x</code> inside it. The struct
<code>x</code> has two fields, <code>y</code> and <code>z</code>, both of which
are 64-bit integers.</td>
</tr>
<tr>
<td style="white-space:nowrap">
<code>
STRUCT&lt;inner_array ARRAY&lt;INT64&gt;&gt;
</code>
</td>
<td>A struct containing an array named <code>inner_array</code> that holds
64-bit integer elements.</td>
</tr>
<tbody>
</table>

### Constructing a struct 
<a id="constructing_a_struct"></a>

#### Tuple syntax

```
(expr1, expr2 [, ... ])
```

The output type is an anonymous struct type with anonymous fields with types
matching the types of the input expressions. There must be at least two
expressions specified. Otherwise this syntax is indistinguishable from an
expression wrapped with parentheses.

**Examples**

<table>
<thead>
<tr>
<th>Syntax</th>
<th>Output Type</th>
<th>Notes</th>
</tr>
</thead>
<tbody>
<tr>
<td style="white-space:nowrap"><code>(x, x+y)</code></td>
<td style="white-space:nowrap"><code>STRUCT&lt;?,?&gt;</code></td>
<td>If column names are used (unquoted strings), the struct field data type is
derived from the column data type. <code>x</code> and <code>y</code> are
columns, so the data types of the struct fields are derived from the column
types and the output type of the addition operator.</td>
</tr>
</tbody>
</table>

This syntax can also be used with struct comparison for comparison expressions
using multi-part keys, e.g., in a `WHERE` clause:

```
WHERE (Key1,Key2) IN ( (12,34), (56,78) )
```

#### Typeless struct syntax

```
STRUCT( expr1 [AS field_name] [, ... ])
```

Duplicate field names are allowed. Fields without names are considered anonymous
fields and can't be referenced by name. struct values can be `NULL`, or can
have `NULL` field values.

**Examples**

<table>
<thead>
<tr>
<th>Syntax</th>
<th>Output Type</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>STRUCT(1,2,3)</code></td>
<td><code>STRUCT&lt;int64,int64,int64&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT()</code></td>
<td><code>STRUCT&lt;&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT('abc')</code></td>
<td><code>STRUCT&lt;string&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT(1, t.str_col)</code></td>
<td><code>STRUCT&lt;int64, str_col string&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT(1 AS a, 'abc' AS b)</code></td>
<td><code>STRUCT&lt;a int64, b string&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT(str_col AS abc)</code></td>
<td><code>STRUCT&lt;abc string&gt;</code></td>
</tr>
</tbody>
</table>

#### Typed struct syntax

```
STRUCT<[field_name] field_type, ...>( expr1 [, ... ])
```

Typed syntax allows constructing structs with an explicit struct data type. The
output type is exactly the `field_type` provided. The input expression is
coerced to `field_type` if the two types aren't the same, and an error is
produced if the types aren't compatible. `AS alias` isn't allowed on the input
expressions. The number of expressions must match the number of fields in the
type, and the expression types must be coercible or literal-coercible to the
field types.

**Examples**

<table>
<thead>
<tr>
<th>Syntax</th>
<th>Output Type</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>STRUCT&lt;int64&gt;(5)</code></td>
<td><code>STRUCT&lt;int64&gt;</code></td>
</tr>

<tr>
<td><code>STRUCT&lt;date&gt;("2011-05-05")</code></td>
<td><code>STRUCT&lt;date&gt;</code></td>
</tr>

<tr>
<td><code>STRUCT&lt;x int64, y string&gt;(1, t.str_col)</code></td>
<td><code>STRUCT&lt;x int64, y string&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT&lt;int64&gt;(int_col)</code></td>
<td><code>STRUCT&lt;int64&gt;</code></td>
</tr>
<tr>
<td><code>STRUCT&lt;x int64&gt;(5 AS x)</code></td>
<td>Error - Typed syntax doesn't allow <code>AS</code></td>
</tr>
</tbody>
</table>

### Limited comparisons for structs

Structs can be directly compared using equality operators:

+ Equal (`=`)
+ Not Equal (`!=` or `<>`)
+ [`NOT`] `IN`

Notice, though, that these direct equality comparisons compare the fields of
the struct pairwise in ordinal order ignoring any field names. If instead you
want to compare identically named fields of a struct, you can compare the
individual fields directly.

[struct-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#struct_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
