---
name: JSON_ARRAY_APPEND
dialect: googlesql
category: functions/json
status: implemented
notes: |
  Recognised by the analyzer but the runtime UDF is not yet registered.  will ship the strict / array / lax variants as one batch under the JSON extractor family.
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_array_append
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_array_append.yaml
---

# JSON_ARRAY_APPEND

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

## `JSON_ARRAY_APPEND`

```googlesql
JSON_ARRAY_APPEND(
  json_expr,
  json_path_value_pair[, ...]
  [, append_each_element => { TRUE | FALSE } ]
)

json_path_value_pair:
  json_path, value
```

Appends JSON data to the end of a JSON array.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '["a", "b", "c"]'
    ```
+   `json_path_value_pair`: A value and the [JSONPath][JSONPath-format] for
    that value. This includes:

    +   `json_path`: Append `value` at this [JSONPath][JSONPath-format]
        in `json_expr`.

    +   `value`: A [JSON encoding-supported][json-encodings] value to
        append.
+   `append_each_element`: A named argument with a `BOOL` value.

    +   If `TRUE` (default), and `value` is a SQL array,
        appends each element individually.

    +   If `FALSE,` and `value` is a SQL array, appends
        the array as one element.

Details:

+   Path value pairs are evaluated left to right. The JSON produced by
    evaluating one pair becomes the JSON against which the next pair
    is evaluated.
+   The operation is ignored if the path points to a JSON non-array value that
    isn't a JSON null.
+   If `json_path` points to a JSON null, the JSON null is replaced by a
    JSON array that contains `value`.
+   If the path exists but has an incompatible type at any given path token,
    the path value pair operation is ignored.
+   The function applies all path value pair append operations even if an
    individual path value pair operation is invalid. For invalid operations,
    the operation is ignored and the function continues to process the rest of
    the path value pairs.
+   If any `json_path` is an invalid [JSONPath][JSONPath-format], an error is
    produced.
+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   If `append_each_element` is SQL `NULL`, the function returns `json_expr`.
+   If `json_path` is SQL `NULL`, the `json_path_value_pair` operation is
    ignored.

**Return type**

`JSON`

**Examples**

In the following example, path `$` is matched and appends `1`.

```googlesql
SELECT JSON_ARRAY_APPEND(JSON '["a", "b", "c"]', '$', 1) AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | ["a","b","c",1] |
 +-----------------*/
```

In the following example, `append_each_element` defaults to `TRUE`, so
`[1, 2]` is appended as individual elements.

```googlesql
SELECT JSON_ARRAY_APPEND(JSON '["a", "b", "c"]', '$', [1, 2]) AS json_data

/*-------------------+
 | json_data         |
 +-------------------+
 | ["a","b","c",1,2] |
 +-------------------*/
```

In the following example, `append_each_element` is `FALSE`, so
`[1, 2]` is appended as one element.

```googlesql
SELECT JSON_ARRAY_APPEND(
  JSON '["a", "b", "c"]',
  '$', [1, 2],
  append_each_element=>FALSE) AS json_data

/*---------------------+
 | json_data           |
 +---------------------+
 | ["a","b","c",[1,2]] |
 +---------------------*/
```

In the following example, `append_each_element` is `FALSE`, so
`[1, 2]` and `[3, 4]` are each appended as one element.

```googlesql
SELECT JSON_ARRAY_APPEND(
  JSON '["a", ["b"], "c"]',
  '$[1]', [1, 2],
  '$[1][1]', [3, 4],
  append_each_element=>FALSE) AS json_data

/*-----------------------------+
 | json_data                   |
 +-----------------------------+
 | ["a",["b",[1,2,[3,4]]],"c"] |
 +-----------------------------*/
```

In the following example, the first path `$[1]` appends `[1, 2]` as single
elements, and then the second path `$[1][1]` isn't a valid path to an array,
so the second operation is ignored.

```googlesql
SELECT JSON_ARRAY_APPEND(
  JSON '["a", ["b"], "c"]',
  '$[1]', [1, 2],
  '$[1][1]', [3, 4]) AS json_data

/*---------------------+
 | json_data           |
 +---------------------+
 | ["a",["b",1,2],"c"] |
 +---------------------*/
```

In the following example, path `$.a` is matched and appends `2`.

```googlesql
SELECT JSON_ARRAY_APPEND(JSON '{"a": [1]}', '$.a', 2) AS json_data

/*-------------+
 | json_data   |
 +-------------+
 | {"a":[1,2]} |
 +-------------*/
```

In the following example, a value is appended into a JSON null.

```googlesql
SELECT JSON_ARRAY_APPEND(JSON '{"a": null}', '$.a', 10)

/*------------+
 | json_data  |
 +------------+
 | {"a":[10]} |
 +------------*/
```

In the following example, path `$.a` isn't an array, so the operation is
ignored.

```googlesql
SELECT JSON_ARRAY_APPEND(JSON '{"a": 1}', '$.a', 2) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":1}   |
 +-----------*/
```

In the following example, path `$.b` doesn't exist, so the operation is
ignored.

```googlesql
SELECT JSON_ARRAY_APPEND(JSON '{"a": 1}', '$.b', 2) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":1}   |
 +-----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
