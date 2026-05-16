---
name: JSON_REMOVE
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_remove
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_remove.yaml
---

# JSON_REMOVE

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

## `JSON_REMOVE`

```googlesql
JSON_REMOVE(json_expr, json_path[, ...])
```

Produces a new SQL `JSON` value with the specified JSON data removed.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '{"class": {"students": [{"name": "Jane"}]}}'
    ```
+   `json_path`: Remove data at this [JSONPath][JSONPath-format] in `json_expr`.

Details:

+   Paths are evaluated left to right. The JSON produced by evaluating the
    first path is the JSON for the next path.
+   The operation ignores non-existent paths and continue processing the rest
    of the paths.
+   For each path, the entire matched JSON subtree is deleted.
+   If the path matches a JSON object key, this function deletes the
    key-value pair.
+   If the path matches an array element, this function deletes the specific
    element from the matched array.
+   If removing the path results in an empty JSON object or empty JSON array,
    the empty structure is preserved.
+   If `json_path` is `$` or an invalid [JSONPath][JSONPath-format], an error is
    produced.
+   If `json_path` is SQL `NULL`, the path operation is ignored.

**Return type**

`JSON`

**Examples**

In the following example, the path `$[1]` is matched and removes
`["b", "c"]`.

```googlesql
SELECT JSON_REMOVE(JSON '["a", ["b", "c"], "d"]', '$[1]') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | ["a","d"] |
 +-----------*/
```

You can use the field access operator to pass JSON data into this function.
For example:

```googlesql
WITH T AS (SELECT JSON '{"a": {"b": 10, "c": 20}}' AS data)
SELECT JSON_REMOVE(data.a, '$.b') AS json_data FROM T

/*-----------+
 | json_data |
 +-----------+
 | {"c":20}  |
 +-----------*/
```

In the following example, the first path `$[1]` is matched and removes
`["b", "c"]`. Then, the second path `$[1]` is matched and removes `"d"`.

```googlesql
SELECT JSON_REMOVE(JSON '["a", ["b", "c"], "d"]', '$[1]', '$[1]') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | ["a"]     |
 +-----------*/
```

The structure of an empty array is preserved when all elements are deleted
from it. For example:

```googlesql
SELECT JSON_REMOVE(JSON '["a", ["b", "c"], "d"]', '$[1]', '$[1]', '$[0]') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | []        |
 +-----------*/
```

In the following example, the path `$.a.b.c` is matched and removes the
`"c":"d"` key-value pair from the JSON object.

```googlesql
SELECT JSON_REMOVE(JSON '{"a": {"b": {"c": "d"}}}', '$.a.b.c') AS json_data

/*----------------+
 | json_data      |
 +----------------+
 | {"a":{"b":{}}} |
 +----------------*/
```

In the following example, the path `$.a.b` is matched and removes the
`"b": {"c":"d"}` key-value pair from the JSON object.

```googlesql
SELECT JSON_REMOVE(JSON '{"a": {"b": {"c": "d"}}}', '$.a.b') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":{}}  |
 +-----------*/
```

In the following example, the path `$.b` isn't valid, so the operation makes
no changes.

```googlesql
SELECT JSON_REMOVE(JSON '{"a": 1}', '$.b') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":1}   |
 +-----------*/
```

In the following example, path `$.a.b` and `$.b` don't exist, so those
operations are ignored, but the others are processed.

```googlesql
SELECT JSON_REMOVE(JSON '{"a": [1, 2, 3]}', '$.a[0]', '$.a.b', '$.b', '$.a[0]') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":[3]} |
 +-----------*/
```

If you pass in `$` as the path, an error is produced. For example:

```googlesql
-- Error: The JSONPath can't be '$'
SELECT JSON_REMOVE(JSON '{}', '$') AS json_data
```

In the following example, the operation is ignored because you can't remove
data from a JSON null.

```googlesql
SELECT JSON_REMOVE(JSON 'null', '$.a.b') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | null      |
 +-----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
