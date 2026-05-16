---
name: JSON_CONTAINS
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_contains
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_contains.yaml
---

# JSON_CONTAINS

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

## `JSON_CONTAINS`

```googlesql
JSON_CONTAINS(json_expr, json_expr)
```

**Description**

Checks if a JSON document contains another JSON document. This function returns
`true` if the first parameter JSON document contains the second parameter JSON
document; otherwise the function returns `false`. If any input argument is
`NULL`, a `NULL` value is returned.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '{"class": {"students": [{"name": "Jane"}]}}'
    ```

Details:

+   The structure and data of the contained document must match a portion of the
    containing document. This function determines if the smaller JSON document
    is part of the larger JSON document.
+   JSON scalars: A JSON scalar value (like a string, number, bool, or JSON null
    ) contains only itself.
+   JSON objects:

    +   An object contains another object if the first object contains all the
        key-value pairs present in the second JSON object.
    +   When checking for object containment, extra key-value pairs in the
        containing object don't prevent a match.
    +   Any JSON object can contain an empty object.
+   JSON arrays:

    +   An array contains another array if every element of the second array is
        contained by some element of the first.
    +   Duplicate elements in arrays are treated as if they appear only once.
    +   The order of elements within JSON arrays isn't significant for
        containment checks.
    +   Any array can contain an empty array.
    +   As a special case, a top-level array can contain a scalar value.

**Return type**

`BOOL`

**Examples**

In the following example, a JSON scalar value (a string) contains only itself:

```googlesql
SELECT JSON_CONTAINS(JSON '"a"', JSON '"a"') AS result;

/*----------+
 |  result  |
 +----------+
 |   true   |
 +----------*/
```

The following examples check if a JSON object contains another JSON object:

```googlesql
SELECT
    JSON_CONTAINS(JSON '{"a": {"b": 1}, "c": 2}', JSON '{"b": 1}') AS result1,
    JSON_CONTAINS(JSON '{"a": {"b": 1}, "c": 2}', JSON '{"a": {"b": 1}}') AS result2,
    JSON_CONTAINS(JSON '{"a": {"b": 1, "d": 3}, "c": 2}', JSON '{"a": {"b": 1}}') AS result3;

/*----------*----------*----------+
 |  result1 |  result2 |  result3 |
 +----------+----------+----------+
 |   false  |   true   |   true   |
 +----------*----------*----------*/
```

The following examples check if a JSON array contains another JSON array. An
array contains another array if the first JSON array contains all the elements
present in the second array. The order of elements doesn't matter.

Also, if the array is a top-level array, it can contain a scalar value.

```googlesql
SELECT
    JSON_CONTAINS(JSON '[1, 2, 3]', JSON '[2]') AS result1,
    JSON_CONTAINS(JSON '[1, 2, 3]', JSON '2') AS result2;

/*----------*----------+
 |  result1 |  result2 |
 +----------+----------+
 |   true   |   true   |
 +----------*----------*/
```

```googlesql
SELECT
    JSON_CONTAINS(JSON '[[1, 2, 3]]', JSON '2') AS result1,
    JSON_CONTAINS(JSON '[[1, 2, 3]]', JSON '[2]') AS result2,
    JSON_CONTAINS(JSON '[[1, 2, 3]]', JSON '[[2]]') AS result3;

/*----------*----------*----------+
 |  result1 |  result2 |  result3 |
 +----------+----------+----------+
 |   false  |   false  |   true   |
 +----------*----------*----------*/
```

The following examples check if a JSON array contains a JSON object:

```googlesql
SELECT
    JSON_CONTAINS(JSON '[{"a":0}, {"b":1, "c":2}]', JSON '[{"b":1}]') AS result1,
    JSON_CONTAINS(JSON '[{"a":0}, {"b":1, "c":2}]', JSON '{"b":1}') AS results2,
    JSON_CONTAINS(JSON '[{"a":0}, {"b":1, "c":2}]', JSON '[{"a":0, "b":1}]') AS results3;

/*----------*----------*----------+
 |  result1 |  result2 |  result3 |
 +----------+----------+----------+
 |   true   |   false  |   false  |
 +----------*----------*----------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
