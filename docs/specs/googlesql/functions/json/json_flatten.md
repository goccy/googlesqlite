---
name: JSON_FLATTEN
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_flatten
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_flatten.yaml
---

# JSON_FLATTEN

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

## `JSON_FLATTEN`

```googlesql
JSON_FLATTEN(json_expr)
```

**Description**

Produces a new SQL `ARRAY<JSON>` value containing all non-array values that are
either directly in the input JSON value or children of one or more consecutively
nested arrays in the input JSON value.

Arguments:

+   `json_expr`: `JSON`. For example:

    ```
    JSON '["Jane", ["John", "Jamie"]]'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.

**Return type**

`ARRAY<JSON>`

**Examples**

In the following example, there is a single non-array value that is returned.

```googlesql
SELECT JSON_FLATTEN(JSON '1') AS json_flatten

/*--------------+
 | json_flatten |
 +--------------+
 | [1]          |
 +--------------*/
```

In the following example, an input array of values is flattened.

```googlesql
SELECT JSON_FLATTEN(JSON '[1, 2, null]') AS json_flatten

/*--------------+
 | json_flatten |
 +--------------+
 | [1, 2, null] |
 +--------------*/
```

In the following example, an input array which includes nested array elements is
flattened.

```googlesql
SELECT JSON_FLATTEN(JSON '[[[1]], 2, [3]]') AS json_flatten

/*--------------+
 | json_flatten |
 +--------------+
 | [1, 2, 3]    |
 +--------------*/
```

In the following example, the nested-array value in a key-value pair is not
flattened because it is enclosed within a JSON object.

```googlesql
SELECT JSON_FLATTEN(JSON '{"a": [[1]]}') AS json_flatten

/*---------------+
 | json_flatten  |
 +---------------+
 | [{"a":[[1]]}] |
 +---------------*/
```

In the following example, the output contains both the flattened array elements
from the input and the non-array elements from the input.

```googlesql
SELECT JSON_FLATTEN(JSON '[[[1, 2], 3], {"a": 4}, true]') AS json_flatten

/*---------------------------+
 | json_flatten              |
 +---------------------------+
 | [1, 2, 3, {"a": 4}, true] |
 +---------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
