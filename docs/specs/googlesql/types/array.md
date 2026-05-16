---
name: ARRAY
dialect: googlesql
category: types
status: implemented
notes: |
  GoogleSQL spec carry-over from earlier sweeps; analyzer / runtime gap. Implementation pending.
source_url: docs/third_party/googlesql-docs/data-types.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/data-types.md#array-type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/types/array.yaml
---

# ARRAY

## Summary

`ARRAY` is an ordered list of zero or more elements, all sharing the same non-array element type.

## Signatures

- `ARRAY<T>` — declared with angle brackets; `T` is any non-array type (structs may wrap a nested array).
- `ARRAY<T>[elem, ...]` — typed array literal.
- `[elem, ...]` — untyped array literal; the element type is inferred from context, defaulting to `ARRAY<INT64>` when no type can be inferred.
- `ARRAY<T>[]` — typed empty array.

## Behavior

- Every element of an array shares a single element type; literals built from mixed-type expressions use the common supertype (for example `INT64` and `DOUBLE` resolve to `DOUBLE`).
- Arrays of arrays are not allowed; nesting must go through a struct, e.g. `ARRAY<STRUCT<ARRAY<INT64>>>`.
- Element types may be arbitrarily complex as long as the immediate element is not itself an array.
- Arrays can be constructed from array literals (`[...]`), typed literals (`ARRAY<T>[...]`), or array-producing functions such as `GENERATE_ARRAY` and `GENERATE_DATE_ARRAY`.
- An array value may itself be `NULL`; inside a query, a `NULL` array and an empty array are distinct values.
- In query results, a `NULL` array is rendered as an empty array, and writing a `NULL` array to a table converts it to an empty array.

## Examples

```googlesql
SELECT [1, 2, 3] AS numbers;
-- expected: [1, 2, 3]
```

```googlesql
SELECT ARRAY<DOUBLE>[1, 2, 3] AS floats;
-- expected: [1.0, 2.0, 3.0]
```

```googlesql
SELECT CAST(NULL AS ARRAY<INT64>) IS NULL AS array_is_null;
-- expected: TRUE
```

```googlesql
SELECT GENERATE_ARRAY(11, 33, 2) AS odds;
-- expected: [11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33]
```

## Edge cases

- A query that would produce an array of arrays raises an error; insert a struct between the arrays (for example via `SELECT AS STRUCT`) instead.
- A `NULL` array and an empty array compare differently inside a query (`IS NULL` is `TRUE` only for the former), but query results render both as `[]`.
- Persisting a `NULL` array to a table silently coerces it to an empty array, so round-tripping loses the `NULL`-ness.
- An untyped empty array literal `[]` whose type cannot be inferred from context defaults to `ARRAY<INT64>`.
- Element expressions with differing types are accepted only when they share a common supertype; otherwise the array literal is rejected.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/data-types.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## Array type 
<a id="array_type"></a>

<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td><code>ARRAY</code></td>
<td>Ordered list of zero or more elements of any non-array type.</td>
</tr>
</tbody>
</table>

An array is an ordered list of zero or more elements of non-array values.
Elements in an array must share the same type.

Arrays of arrays aren't allowed. Queries that would produce an array of
arrays return an error. Instead, a struct must be inserted between the
arrays using the `SELECT AS STRUCT` construct.

To learn more about the literal representation of an array type,
see [Array literals][array-literals].

To learn more about using arrays in GoogleSQL, see [Work with
arrays][working-with-arrays].

### `NULL`s and the array type 
<a id="array_nulls"></a>

Currently, GoogleSQL has the following rules with respect to `NULL`s and
arrays:

+ An array can be `NULL`.

  For example:

  ```googlesql
  SELECT CAST(NULL AS ARRAY<INT64>) IS NULL AS array_is_null;

  /*---------------+
   | array_is_null |
   +---------------+
   | TRUE          |
   +---------------*/
  ```
+ GoogleSQL translates a `NULL` array into an empty array in the query
  result, although inside the query, `NULL` and empty arrays are two distinct
  values.

  For example:

  ```googlesql
  WITH Items AS (
    SELECT [] AS numbers, "Empty array in query" AS description UNION ALL
    SELECT CAST(NULL AS ARRAY<INT64>), "NULL array in query")
  SELECT numbers, description, numbers IS NULL AS numbers_null
  FROM Items;

  /*---------+----------------------+--------------+
   | numbers | description          | numbers_null |
   +---------+----------------------+--------------+
   | []      | Empty array in query | false        |
   | []      | NULL array in query  | true         |
   +---------+----------------------+--------------*/
  ```

  When you write a `NULL` array to a table, it's converted to an
  empty array. If you write `Items` to a table from the previous query,
  then each array is written as an empty array:

  ```googlesql
  SELECT numbers, description, numbers IS NULL AS numbers_null
  FROM Items;

  /*---------+----------------------+--------------+
   | numbers | description          | numbers_null |
   +---------+----------------------+--------------+
   | []      | Empty array in query | false        |
   | []      | NULL array in query  | false        |
   +---------+----------------------+--------------*/
  ```

### Declaring an array type

```
ARRAY<T>
```

Array types are declared using the angle brackets (`<` and `>`). The type
of the elements of an array can be arbitrarily complex with the exception that
an array can't directly contain another array.

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
ARRAY&lt;INT64&gt;
</code>
</td>
<td>Simple array of 64-bit integers.</td>
</tr>
<tr>
<td style="white-space:nowrap">
<code>
ARRAY&lt;STRUCT&lt;INT64, INT64&gt;&gt;
</code>
</td>
<td>An array of structs, each of which contains two 64-bit integers.</td>
</tr>
<tr>
<td style="white-space:nowrap">
<code>
ARRAY&lt;ARRAY&lt;INT64&gt;&gt;
</code><br/>
(not supported)
</td>
<td>This is an <strong>invalid</strong> type declaration which is included here
just in case you came looking for how to create a multi-level array. Arrays
can't contain arrays directly. Instead see the next example.</td>
</tr>
<tr>
<td style="white-space:nowrap">
<code>
ARRAY&lt;STRUCT&lt;ARRAY&lt;INT64&gt;&gt;&gt;
</code>
</td>
<td>An array of arrays of 64-bit integers. Notice that there is a struct between
the two arrays because arrays can't hold other arrays directly.</td>
</tr>
<tbody>
</table>

### Constructing an array 
<a id="constructing_an_array"></a>

You can construct an array using array literals or array functions.

#### Using array literals

You can build an array literal in GoogleSQL using brackets (`[` and
`]`). Each element in an array is separated by a comma.

```googlesql
SELECT [1, 2, 3] AS numbers;

SELECT ["apple", "pear", "orange"] AS fruit;

SELECT [true, false, true] AS booleans;
```

You can also create arrays from any expressions that have compatible types. For
example:

```googlesql
SELECT [a, b, c]
FROM
  (SELECT 5 AS a,
          37 AS b,
          406 AS c);

SELECT [a, b, c]
FROM
  (SELECT CAST(5 AS INT64) AS a,
          CAST(37 AS DOUBLE) AS b,
          406 AS c);
```

Notice that the second example contains three expressions: one that returns an
`INT64`, one that returns a `DOUBLE`, and one that
declares a literal. This expression works because all three expressions share
`DOUBLE` as a supertype.

To declare a specific data type for an array, use angle
brackets (`<` and `>`). For example:

```googlesql
SELECT ARRAY<DOUBLE>[1, 2, 3] AS floats;
```

Arrays of most data types, such as `INT64` or `STRING`, don't require
that you declare them first.

```googlesql
SELECT [1, 2, 3] AS numbers;
```

You can write an empty array of a specific type using `ARRAY<type>[]`. You can
also write an untyped empty array using `[]`, in which case GoogleSQL
attempts to infer the array type from the surrounding context. If
GoogleSQL can't infer a type, the default type `ARRAY<INT64>` is used.

#### Using generated values

You can also construct an `ARRAY` with generated values.

##### Generating arrays of integers

[`GENERATE_ARRAY`][generate-array-function]
generates an array of values from a starting and ending value and a step value.
For example, the following query generates an array that contains all of the odd
integers from 11 to 33, inclusive:

```googlesql
SELECT GENERATE_ARRAY(11, 33, 2) AS odds;

/*--------------------------------------------------+
 | odds                                             |
 +--------------------------------------------------+
 | [11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33] |
 +--------------------------------------------------*/
```

You can also generate an array of values in descending order by giving a
negative step value:

```googlesql
SELECT GENERATE_ARRAY(21, 14, -1) AS countdown;

/*----------------------------------+
 | countdown                        |
 +----------------------------------+
 | [21, 20, 19, 18, 17, 16, 15, 14] |
 +----------------------------------*/
```

##### Generating arrays of dates

[`GENERATE_DATE_ARRAY`][generate-date-array]
generates an array of `DATE`s from a starting and ending `DATE` and a step
`INTERVAL`.

You can generate a set of `DATE` values using `GENERATE_DATE_ARRAY`. For
example, this query returns the current `DATE` and the following
`DATE`s at 1 `WEEK` intervals up to and including a later `DATE`:

```googlesql
SELECT
  GENERATE_DATE_ARRAY('2017-11-21', '2017-12-31', INTERVAL 1 WEEK)
    AS date_array;

/*--------------------------------------------------------------------------+
 | date_array                                                               |
 +--------------------------------------------------------------------------+
 | [2017-11-21, 2017-11-28, 2017-12-05, 2017-12-12, 2017-12-19, 2017-12-26] |
 +--------------------------------------------------------------------------*/
```

[array-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#array_literals

[working-with-arrays]: https://github.com/google/googlesql/blob/master/docs/arrays.md#constructing_arrays

[generate-array-function]: https://github.com/google/googlesql/blob/master/docs/array_functions.md#generate_array

[generate-date-array]: https://github.com/google/googlesql/blob/master/docs/array_functions.md#generate_date_array

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/data-types.md`.
