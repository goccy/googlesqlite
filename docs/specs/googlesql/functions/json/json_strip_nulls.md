---
name: JSON_STRIP_NULLS
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_strip_nulls
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_strip_nulls.yaml
---

# JSON_STRIP_NULLS

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

## `JSON_STRIP_NULLS`

```googlesql
JSON_STRIP_NULLS(
  json_expr
  [, json_path ]
  [, include_arrays => { TRUE | FALSE } ]
  [, remove_empty => { TRUE | FALSE } ]
)
```

Recursively removes JSON nulls from JSON objects and JSON arrays.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '{"a": null, "b": "c"}'
    ```
+   `json_path`: Remove JSON nulls at this [JSONPath][JSONPath-format] for
    `json_expr`.
+   `include_arrays`: A named argument that's either
     `TRUE` (default) or `FALSE`. If `TRUE` or omitted, the function removes
     JSON nulls from JSON arrays. If `FALSE`, doesn't.
+   `remove_empty`: A named argument that's either
     `TRUE` or `FALSE` (default). If `TRUE`, the function removes empty
     JSON objects after JSON nulls are removed. If `FALSE` or omitted, doesn't.

    If `remove_empty` is `TRUE` and `include_arrays` is `TRUE` or omitted,
    the function additionally removes empty JSON arrays.

Details:

+   If a value is a JSON null, the associated key-value pair is removed.
+   If `remove_empty` is set to `TRUE`, the function recursively removes empty
    containers after JSON nulls are removed.
+   If the function generates JSON with nothing in it, the function returns a
    JSON null.
+   If `json_path` is an invalid [JSONPath][JSONPath-format], an error is
    produced.
+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   If `json_path`, `include_arrays`, or `remove_empty` is SQL `NULL`, the
    function returns `json_expr`.

**Return type**

`JSON`

**Examples**

In the following example, all JSON nulls are removed.

```googlesql
SELECT JSON_STRIP_NULLS(JSON '{"a": null, "b": "c"}') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"b":"c"} |
 +-----------*/
```

In the following example, all JSON nulls are removed from a JSON array.

```googlesql
SELECT JSON_STRIP_NULLS(JSON '[1, null, 2, null]') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | [1,2]     |
 +-----------*/
```

In the following example, `include_arrays` is set as `FALSE` so that JSON nulls
aren't removed from JSON arrays.

```googlesql
SELECT JSON_STRIP_NULLS(JSON '[1, null, 2, null]', include_arrays=>FALSE) AS json_data

/*-----------------+
 | json_data       |
 +-----------------+
 | [1,null,2,null] |
 +-----------------*/
```

In the following example, `remove_empty` is omitted and defaults to
`FALSE`, and the empty structures are retained.

```googlesql
SELECT JSON_STRIP_NULLS(JSON '[1, null, 2, null, [null]]') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | [1,2,[]]  |
 +-----------*/
```

In the following example, `remove_empty` is set as `TRUE`, and the
empty structures are removed.

```googlesql
SELECT JSON_STRIP_NULLS(
  JSON '[1, null, 2, null, [null]]',
  remove_empty=>TRUE) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | [1,2]     |
 +-----------*/
```

In the following examples, `remove_empty` is set as `TRUE`, and the
empty structures are removed. Because no JSON data is left the function
returns JSON null.

```googlesql
SELECT JSON_STRIP_NULLS(JSON '{"a": null}', remove_empty=>TRUE) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | null      |
 +-----------*/
```

```googlesql
SELECT JSON_STRIP_NULLS(JSON '{"a": [null]}', remove_empty=>TRUE) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | null      |
 +-----------*/
```

In the following example, empty structures are removed for JSON objects,
but not JSON arrays.

```googlesql
SELECT JSON_STRIP_NULLS(
  JSON '{"a": {"b": {"c": null}}, "d": [null], "e": [], "f": 1}',
  include_arrays=>FALSE,
  remove_empty=>TRUE) AS json_data

/*---------------------------+
 | json_data                 |
 +---------------------------+
 | {"d":[null],"e":[],"f":1} |
 +---------------------------*/
```

In the following example, empty structures are removed for both JSON objects,
and JSON arrays.

```googlesql
SELECT JSON_STRIP_NULLS(
  JSON '{"a": {"b": {"c": null}}, "d": [null], "e": [], "f": 1}',
  remove_empty=>TRUE) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"f":1}   |
 +-----------*/
```

In the following example, because no JSON data is left, the function returns a
JSON null.

```googlesql
SELECT JSON_STRIP_NULLS(JSON 'null') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | null      |
 +-----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
