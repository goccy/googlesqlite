---
name: STRING_AGG
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#string_agg
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/string_agg.yaml
---

# STRING_AGG

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

Verbatim copy from `docs/third_party/googlesql-docs/aggregate_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `STRING_AGG`

```googlesql
STRING_AGG(
  [ DISTINCT ]
  expression [, delimiter]
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
  [ ORDER BY key [ { ASC | DESC } ] [, ... ] ]
  [ LIMIT n ]
)
[ OVER over_clause ]

over_clause:
  { named_window | ( [ window_specification ] ) }

window_specification:
  [ named_window ]
  [ PARTITION BY partition_expression [, ...] ]
  [ ORDER BY expression [ { ASC | DESC }  ] [, ...] ]
  [ window_frame_clause ]

```

**Description**

Returns a value (either `STRING` or `BYTES`) obtained by concatenating
non-`NULL` values. Returns `NULL` if there are zero input rows or `expression`
evaluates to `NULL` for all rows.

If a `delimiter` is specified, concatenated values are separated by that
delimiter; otherwise, a comma is used as a delimiter.

To learn more about the optional aggregate clauses that you can pass
into this function, see
[Aggregate function calls][aggregate-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[aggregate-function-calls]: https://github.com/google/googlesql/blob/master/docs/aggregate-function-calls.md

<!-- mdlint on -->

To learn more about the `OVER` clause and how to use it, see
[Window function calls][window-function-calls].

<!-- mdlint off(WHITESPACE_LINE_LENGTH) -->

[window-function-calls]: https://github.com/google/googlesql/blob/master/docs/window-function-calls.md

<!-- mdlint on -->

**Supported Argument Types**

Either `STRING` or `BYTES`.

**Return Data Types**

Either `STRING` or `BYTES`.

**Examples**

```googlesql
SELECT STRING_AGG(fruit) AS string_agg
FROM UNNEST(["apple", NULL, "pear", "banana", "pear"]) AS fruit;

/*------------------------+
 | string_agg             |
 +------------------------+
 | apple,pear,banana,pear |
 +------------------------*/
```

```googlesql
SELECT STRING_AGG(fruit, " & ") AS string_agg
FROM UNNEST(["apple", "pear", "banana", "pear"]) AS fruit;

/*------------------------------+
 | string_agg                   |
 +------------------------------+
 | apple & pear & banana & pear |
 +------------------------------*/
```

```googlesql
SELECT STRING_AGG(DISTINCT fruit, " & ") AS string_agg
FROM UNNEST(["apple", "pear", "banana", "pear"]) AS fruit;

/*-----------------------+
 | string_agg            |
 +-----------------------+
 | apple & pear & banana |
 +-----------------------*/
```

```googlesql
SELECT STRING_AGG(fruit, " & " ORDER BY LENGTH(fruit)) AS string_agg
FROM UNNEST(["apple", "pear", "banana", "pear"]) AS fruit;

/*------------------------------+
 | string_agg                   |
 +------------------------------+
 | pear & pear & apple & banana |
 +------------------------------*/
```

```googlesql
SELECT STRING_AGG(fruit, " & " LIMIT 2) AS string_agg
FROM UNNEST(["apple", "pear", "banana", "pear"]) AS fruit;

/*--------------+
 | string_agg   |
 +--------------+
 | apple & pear |
 +--------------*/
```

```googlesql
SELECT STRING_AGG(DISTINCT fruit, " & " ORDER BY fruit DESC LIMIT 2) AS string_agg
FROM UNNEST(["apple", "pear", "banana", "pear"]) AS fruit;

/*---------------+
 | string_agg    |
 +---------------+
 | pear & banana |
 +---------------*/
```

```googlesql
SELECT
  fruit,
  STRING_AGG(fruit, " & ") OVER (ORDER BY LENGTH(fruit)) AS string_agg
FROM UNNEST(["apple", NULL, "pear", "banana", "pear"]) AS fruit;

/*--------+------------------------------+
 | fruit  | string_agg                   |
 +--------+------------------------------+
 | NULL   | NULL                         |
 | pear   | pear & pear                  |
 | pear   | pear & pear                  |
 | apple  | pear & pear & apple          |
 | banana | pear & pear & apple & banana |
 +--------+------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
