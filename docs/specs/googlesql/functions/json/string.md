---
name: STRING
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/string.yaml
---

# STRING

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

## `STRING` 
<a id="string_for_json"></a>

```googlesql
STRING(json_expr)
```

**Description**

Converts a JSON string to a SQL `STRING` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '"purple"'
    ```

    If the JSON value isn't a string, an error is produced. If the expression
    is SQL `NULL`, the function returns SQL `NULL`.

**Return type**

`STRING`

**Examples**

```googlesql
SELECT STRING(JSON '"purple"') AS color;

/*--------+
 | color  |
 +--------+
 | purple |
 +--------*/
```

```googlesql
SELECT STRING(JSON_QUERY(JSON '{"name": "sky", "color": "blue"}', "$.color")) AS color;

/*-------+
 | color |
 +-------+
 | blue  |
 +-------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if the JSON isn't of type string.
SELECT STRING(JSON '123') AS result; -- Throws an error
SELECT STRING(JSON 'null') AS result; -- Throws an error
SELECT SAFE.STRING(JSON '123') AS result; -- Returns a SQL NULL
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
