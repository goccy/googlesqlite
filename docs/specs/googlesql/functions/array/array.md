---
name: ARRAY
dialect: googlesql
category: functions/array
status: implemented
source_url: docs/third_party/googlesql-docs/array_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/array_functions.md#array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/array/array.yaml
---

# ARRAY

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

Verbatim copy from `docs/third_party/googlesql-docs/array_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ARRAY`

```googlesql
ARRAY(subquery)
```

**Description**

The `ARRAY` function returns an `ARRAY` with one element for each row in a
[subquery][subqueries].

If `subquery` produces a
[SQL table][datamodel-sql-tables],
the table must have exactly one column. Each element in the output `ARRAY` is
the value of the single column of a row in the table.

If `subquery` produces a
[value table][datamodel-value-tables],
then each element in the output `ARRAY` is the entire corresponding row of the
value table.

**Constraints**

+ Subqueries are unordered, so the elements of the output `ARRAY` aren't
guaranteed to preserve any order in the source table for the subquery. However,
if the subquery includes an `ORDER BY` clause, the `ARRAY` function will return
an `ARRAY` that honors that clause.
+ If the subquery returns more than one column, the `ARRAY` function returns an
error.
+ If the subquery returns an `ARRAY` typed column or `ARRAY` typed rows, the
  `ARRAY` function returns an error that GoogleSQL doesn't support
  `ARRAY`s with elements of type
  [`ARRAY`][array-data-type].
+ If the subquery returns zero rows, the `ARRAY` function returns an empty
`ARRAY`. It never returns a `NULL` `ARRAY`.

**Return type**

`ARRAY`

**Examples**

```googlesql
SELECT ARRAY
  (SELECT 1 UNION ALL
   SELECT 2 UNION ALL
   SELECT 3) AS new_array;

/*-----------+
 | new_array |
 +-----------+
 | [1, 2, 3] |
 +-----------*/
```

To construct an `ARRAY` from a subquery that contains multiple
columns, change the subquery to use `SELECT AS STRUCT`. Now
the `ARRAY` function will return an `ARRAY` of `STRUCT`s. The `ARRAY` will
contain one `STRUCT` for each row in the subquery, and each of these `STRUCT`s
will contain a field for each column in that row.

```googlesql
SELECT
  ARRAY
    (SELECT AS STRUCT 1, 2, 3
     UNION ALL SELECT AS STRUCT 4, 5, 6) AS new_array;

/*------------------------+
 | new_array              |
 +------------------------+
 | [{1, 2, 3}, {4, 5, 6}] |
 +------------------------*/
```

Similarly, to construct an `ARRAY` from a subquery that contains
one or more `ARRAY`s, change the subquery to use `SELECT AS STRUCT`.

```googlesql
SELECT ARRAY
  (SELECT AS STRUCT [1, 2, 3] UNION ALL
   SELECT AS STRUCT [4, 5, 6]) AS new_array;

/*----------------------------+
 | new_array                  |
 +----------------------------+
 | [{[1, 2, 3]}, {[4, 5, 6]}] |
 +----------------------------*/
```

[subqueries]: https://github.com/google/googlesql/blob/master/docs/subqueries.md

[datamodel-sql-tables]: https://github.com/google/googlesql/blob/master/docs/data-model.md#standard_sql_tables

[datamodel-value-tables]: https://github.com/google/googlesql/blob/master/docs/data-model.md#value_tables

[array-data-type]: https://github.com/google/googlesql/blob/master/docs/data-types.md#array_type

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/array_functions.md`.
