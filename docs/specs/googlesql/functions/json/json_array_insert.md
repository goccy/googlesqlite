---
name: JSON_ARRAY_INSERT
dialect: googlesql
category: functions/json
status: implemented
notes: |
  json_array_insert is shipped by SQLite but needs odd-argument signature support that our adapter does not yet route through;  will add the variadic path.
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_array_insert
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_array_insert.yaml
---

# JSON_ARRAY_INSERT

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

Verbatim copy from `docs/third_party/googlesql-docs/json_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `JSON_ARRAY_INSERT`

```googlesql
JSON_ARRAY_INSERT(
  json_expr,
  json_path_value_pair[, ...]
  [, insert_each_element => { TRUE | FALSE } ]
)

json_path_value_pair:
  json_path, value
```

Produces a new JSON value that's created by inserting JSON data into
a JSON array.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '["a", "b", "c"]'
    ```
+   `json_path_value_pair`: A value and the [JSONPath][JSONPath-format] for
    that value. This includes:

    +   `json_path`: Insert `value` at this [JSONPath][JSONPath-format]
        in `json_expr`.

    +   `value`: A [JSON encoding-supported][json-encodings] value to
        insert.
+   `insert_each_element`: A named argument with a `BOOL` value.

    +   If `TRUE` (default), and `value` is a SQL array,
        inserts each element individually.

    +   If `FALSE,` and `value` is a SQL array, inserts
        the array as one element.

Details:

+   Path value pairs are evaluated left to right. The JSON produced by
    evaluating one pair becomes the JSON against which the next pair
    is evaluated.
+   The operation is ignored if the path points to a JSON non-array value that
    isn't a JSON null.
+   If `json_path` points to a JSON null, the JSON null is replaced by a
    JSON array of the appropriate size and padded on the left with JSON nulls.
+   If the path exists but has an incompatible type at any given path token,
    the path value pair operator is ignored.
+   The function applies all path value pair append operations even if an
    individual path value pair operation is invalid. For invalid operations,
    the operation is ignored and the function continues to process the rest of
    the path value pairs.
+   If the array index in `json_path` is larger than the size of the array, the
    function extends the length of the array to the index, fills in
    the array with JSON nulls, then adds `value` at the index.
+   If any `json_path` is an invalid [JSONPath][JSONPath-format], an error is
    produced.
+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   If `insert_each_element` is SQL `NULL`, the function returns `json_expr`.
+   If `json_path` is SQL `NULL`, the `json_path_value_pair` operation is
    ignored.

**Return type**

`JSON`

**Examples**

In the following example, path `$[1]` is matched and inserts `1`.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '["a", ["b", "c"], "d"]', '$[1]', 1) AS json_data

/*-----------------------+
 | json_data             |
 +-----------------------+
 | ["a",1,["b","c"],"d"] |
 +-----------------------*/
```

In the following example, path `$[1][0]` is matched and inserts `1`.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '["a", ["b", "c"], "d"]', '$[1][0]', 1) AS json_data

/*-----------------------+
 | json_data             |
 +-----------------------+
 | ["a",[1,"b","c"],"d"] |
 +-----------------------*/
```

In the following example, `insert_each_element` defaults to `TRUE`, so
`[1, 2]` is inserted as individual elements.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '["a", "b", "c"]', '$[1]', [1, 2]) AS json_data

/*-------------------+
 | json_data         |
 +-------------------+
 | ["a",1,2,"b","c"] |
 +-------------------*/
```

In the following example, `insert_each_element` is `FALSE`, so `[1, 2]` is
inserted as one element.

```googlesql
SELECT JSON_ARRAY_INSERT(
  JSON '["a", "b", "c"]',
  '$[1]', [1, 2],
  insert_each_element=>FALSE) AS json_data

/*---------------------+
 | json_data           |
 +---------------------+
 | ["a",[1,2],"b","c"] |
 +---------------------*/
```

In the following example, path `$[7]` is larger than the length of the
matched array, so the array is extended with JSON nulls and `"e"` is inserted at
the end of the array.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '["a", "b", "c", "d"]', '$[7]', "e") AS json_data

/*--------------------------------------+
 | json_data                            |
 +--------------------------------------+
 | ["a","b","c","d",null,null,null,"e"] |
 +--------------------------------------*/
```

In the following example, path `$.a` is an object, so the operation is ignored.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '{"a": {}}', '$.a[0]', 2) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":{}}  |
 +-----------*/
```

In the following example, path `$` doesn't specify a valid array position,
so the operation is ignored.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '[1, 2]', '$', 3) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | [1,2]     |
 +-----------*/
```

In the following example, a value is inserted into a JSON null.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '{"a": null}', '$.a[2]', 10) AS json_data

/*----------------------+
 | json_data            |
 +----------------------+
 | {"a":[null,null,10]} |
 +----------------------*/
```

In the following example, the operation is ignored because you can't insert
data into a JSON number.

```googlesql
SELECT JSON_ARRAY_INSERT(JSON '1', '$[0]', 'r1') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | 1         |
 +-----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
