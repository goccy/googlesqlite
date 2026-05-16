---
name: MAX
dialect: googlesql
category: functions/aggregate
status: implemented
source_url: docs/third_party/googlesql-docs/aggregate_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/aggregate_functions.md#max
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/aggregate/max.yaml
---

# MAX

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

## `MAX`

```googlesql
MAX(
  expression
  [ WHERE where_expression ]
  [ HAVING { MAX | MIN } having_expression ]
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

Returns the maximum non-`NULL` value in an aggregated group.

Caveats:

+ If the aggregated group is empty or the argument is `NULL` for all rows in
  the group, returns `NULL`.
+ If the argument is `NaN` for any row in the group, returns `NaN`.

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

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Supported Argument Types**

Any [orderable data type][agg-data-type-properties] except for `ARRAY`.

**Return Data Types**

The data type of the input values.

**Examples**

```googlesql
SELECT MAX(x) AS max
FROM UNNEST([8, 37, 55, 4]) AS x;

/*-----+
 | max |
 +-----+
 | 55  |
 +-----*/
```

```googlesql
SELECT x, MAX(x) OVER (PARTITION BY MOD(x, 2)) AS max
FROM UNNEST([8, NULL, 37, 55, NULL, 4]) AS x;

/*------+------+
 | x    | max  |
 +------+------+
 | NULL | NULL |
 | NULL | NULL |
 | 8    | 8    |
 | 4    | 8    |
 | 37   | 55   |
 | 55   | 55   |
 +------+------*/
```

[agg-data-type-properties]: https://github.com/google/googlesql/blob/master/docs/data-types.md#data_type_properties

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/aggregate_functions.md`.
