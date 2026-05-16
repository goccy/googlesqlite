---
name: STRING_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#string_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/string_array.yaml
---

# STRING_ARRAY

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

## `STRING_ARRAY` 
<a id="string_array_for_json"></a>

```googlesql
STRING_ARRAY(json_expr)
```

**Description**

Converts a JSON array of strings to a SQL `ARRAY<STRING>` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '["purple", "blue"]'
    ```

    If the JSON value isn't an array of strings, an error is produced. If the
    expression is SQL `NULL`, the function returns SQL `NULL`.

**Return type**

`ARRAY<STRING>`

**Examples**

```googlesql
SELECT STRING_ARRAY(JSON '["purple", "blue"]') AS colors;

/*----------------+
 | colors         |
 +----------------+
 | [purple, blue] |
 +----------------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if the JSON isn't an array of strings.
SELECT STRING_ARRAY(JSON '[123]') AS result; -- Throws an error
SELECT STRING_ARRAY(JSON '[null]') AS result; -- Throws an error
SELECT STRING_ARRAY(JSON 'null') AS result; -- Throws an error
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
