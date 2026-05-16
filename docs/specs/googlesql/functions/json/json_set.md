---
name: JSON_SET
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_set
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_set.yaml
---

# JSON_SET

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

## `JSON_SET`

```googlesql
JSON_SET(
  json_expr,
  json_path_value_pair[, ...]
  [, create_if_missing => { TRUE | FALSE } ]
)

json_path_value_pair:
  json_path, value
```

Produces a new SQL `JSON` value with the specified JSON data inserted
or replaced.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '{"class": {"students": [{"name": "Jane"}]}}'
    ```
+   `json_path_value_pair`: A value and the [JSONPath][JSONPath-format] for
    that value. This includes:

    +   `json_path`: Insert or replace `value` at this [JSONPath][JSONPath-format]
        in `json_expr`.

    +   `value`: A [JSON encoding-supported][json-encodings] value to
        insert.
+   `create_if_missing`: A named argument that takes a `BOOL` value.

    +   If `TRUE` (default), replaces or inserts data if the path doesn't exist.

    +   If `FALSE`, only existing JSONPath values are replaced. If the path
        doesn't exist, the set operation is ignored.

Details:

+   Path value pairs are evaluated left to right. The JSON produced by
    evaluating one pair becomes the JSON against which the next pair
    is evaluated.
+   If a matched path has an existing value, it overwrites the existing data
    with `value`.
+   If `create_if_missing` is `TRUE`:

      +  If a path doesn't exist, the remainder of the path is recursively
         created.
      +  If the matched path prefix points to a JSON null, the remainder of the
         path is recursively created, and `value` is inserted.
      +  If a path token points to a JSON array and the specified index is
         _larger_ than the size of the array, pads the JSON array with JSON
         nulls, recursively creates the remainder of the path at the specified
         index, and inserts the path value pair.
+   This function applies all path value pair set operations even if an
    individual path value pair operation is invalid. For invalid operations,
    the operation is ignored and the function continues to process the rest
    of the path value pairs.
+   If the path exists but has an incompatible type at any given path
    token, no update happens for that specific path value pair.
+   If any `json_path` is an invalid [JSONPath][JSONPath-format], an error is
    produced.
+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   If `json_path` is SQL `NULL`, the `json_path_value_pair` operation is
    ignored.
+   If `create_if_missing` is SQL `NULL`, the set operation is ignored.

**Return type**

`JSON`

**Examples**

In the following example, the path `$` matches the entire `JSON` value
and replaces it with `{"b": 2, "c": 3}`.

```googlesql
SELECT JSON_SET(JSON '{"a": 1}', '$', JSON '{"b": 2, "c": 3}') AS json_data

/*---------------+
 | json_data     |
 +---------------+
 | {"b":2,"c":3} |
 +---------------*/
```

In the following example, `create_if_missing` is `FALSE` and the path `$.b`
doesn't exist, so the set operation is ignored.

```googlesql
SELECT JSON_SET(
  JSON '{"a": 1}',
  "$.b", 999,
  create_if_missing => false) AS json_data

/*------------+
 | json_data  |
 +------------+
 | '{"a": 1}' |
 +------------*/
```

In the following example, `create_if_missing` is `TRUE` and the path `$.a`
exists, so the value is replaced.

```googlesql
SELECT JSON_SET(
  JSON '{"a": 1}',
  "$.a", 999,
  create_if_missing => false) AS json_data

/*--------------+
 | json_data    |
 +--------------+
 | '{"a": 999}' |
 +--------------*/
```

In the following example, the path `$.a` is matched, but `$.a.b` doesn't
exist, so the new path and the value are inserted.

```googlesql
SELECT JSON_SET(JSON '{"a": {}}', '$.a.b', 100) AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | {"a":{"b":100}} |
 +-----------------*/
```

In the following example, the path prefix `$` points to a JSON null, so the
remainder of the path is created for the value `100`.

```googlesql
SELECT JSON_SET(JSON 'null', '$.a.b', 100) AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | {"a":{"b":100}} |
 +-----------------*/
```

In the following example, the path `$.a.c` implies that the value at `$.a` is
a JSON object but it's not. This part of the operation is ignored, but the other
parts of the operation are completed successfully.

```googlesql
SELECT JSON_SET(
  JSON '{"a": 1}',
  '$.b', 2,
  '$.a.c', 100,
  '$.d', 3) AS json_data

/*---------------------+
 | json_data           |
 +---------------------+
 | {"a":1,"b":2,"d":3} |
 +---------------------*/
```

In the following example, the path `$.a[2]` implies that the value for `$.a` is
an array, but it's not, so the operation is ignored for that value.

```googlesql
SELECT JSON_SET(
  JSON '{"a": 1}',
  '$.a[2]', 100,
  '$.b', 2) AS json_data

/*---------------+
 | json_data     |
 +---------------+
 | {"a":1,"b":2} |
 +---------------*/
```

In the following example, the path `$[1]` is matched and replaces the
array element value with `foo`.

```googlesql
SELECT JSON_SET(JSON '["a", ["b", "c"], "d"]', '$[1]', "foo") AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | ["a","foo","d"] |
 +-----------------*/
```

In the following example, the path `$[1][0]` is matched and replaces the
array element value with `foo`.

```googlesql
SELECT JSON_SET(JSON '["a", ["b", "c"], "d"]', '$[1][0]', "foo") AS json_data

/*-----------------------+
 | json_data             |
 +-----------------------+
 | ["a",["foo","c"],"d"] |
 +-----------------------*/
```

In the following example, the path prefix `$` points to a JSON null, so the
remainder of the path is created. The resulting array is padded with
JSON nulls and appended with `foo`.

```googlesql
SELECT JSON_SET(JSON 'null', '$[0][3]', "foo")

/*--------------------------+
 | json_data                |
 +--------------------------+
 | [[null,null,null,"foo"]] |
 +--------------------------*/
```

In the following example, the path `$[1]` is matched, the matched array is
extended since `$[1][4]` is larger than the existing array, and then `foo` is
inserted in the array.

```googlesql
SELECT JSON_SET(JSON '["a", ["b", "c"], "d"]', '$[1][4]', "foo") AS json_data

/*-------------------------------------+
 | json_data                           |
 +-------------------------------------+
 | ["a",["b","c",null,null,"foo"],"d"] |
 +-------------------------------------*/
```

In the following example, the path `$[1][0][0]` implies that the value of
`$[1][0]` is an array, but it isn't, so the operation is ignored.

```googlesql
SELECT JSON_SET(JSON '["a", ["b", "c"], "d"]', '$[1][0][0]', "foo") AS json_data

/*---------------------+
 | json_data           |
 +---------------------+
 | ["a",["b","c"],"d"] |
 +---------------------*/
```

In the following example, the path `$[1][2]` is larger than the length of
the matched array. The array length is extended and the remainder of the path
is recursively created. The operation continues to the path `$[1][2][1]`
and inserts `foo`.

```googlesql
SELECT JSON_SET(JSON '["a", ["b", "c"], "d"]', '$[1][2][1]', "foo") AS json_data

/*----------------------------------+
 | json_data                        |
 +----------------------------------+
 | ["a",["b","c",[null,"foo"]],"d"] |
 +----------------------------------*/
```

In the following example, because the `JSON` object is empty, key `b` is
inserted, and the remainder of the path is recursively created.

```googlesql
SELECT JSON_SET(JSON '{}', '$.b[2].d', 100) AS json_data

/*-----------------------------+
 | json_data                   |
 +-----------------------------+
 | {"b":[null,null,{"d":100}]} |
 +-----------------------------*/
```

In the following example, multiple values are set.

```googlesql
SELECT JSON_SET(
  JSON '{"a": 1, "b": {"c":3}, "d": [4]}',
  '$.a', 'v1',
  '$.b.e', 'v2',
  '$.d[2]', 'v3') AS json_data

/*---------------------------------------------------+
 | json_data                                         |
 +---------------------------------------------------+
 | {"a":"v1","b":{"c":3,"e":"v2"},"d":[4,null,"v3"]} |
 +---------------------------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
