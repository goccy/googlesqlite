---
name: JSON_VALUE
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_value
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_value.yaml
---

# JSON_VALUE

## Summary

Extracts a JSON scalar value at a given JSONPath and converts it to a SQL `STRING`, removing the outermost quotes and unescaping the value.

## Signatures

- `JSON_VALUE(json_string_expr[, json_path])`
- `JSON_VALUE(json_expr[, json_path])`

## Behavior

- Returns `STRING`.
- Accepts either a JSON-formatted string (`json_string_expr`) or a `JSON` value (`json_expr`) as input.
- Removes the outermost quotes from the matched scalar and unescapes the value before returning it.
- When `json_path` is omitted, the JSONPath `$` is applied so the entire input is analyzed.
- Returns SQL `NULL` when the selected value is non-scalar (an object or an array) or is JSON `null`.
- JSON keys that contain invalid JSONPath characters can be escaped in the path using double quotes, for example `$."a.b".c`.

## Examples

```googlesql
SELECT JSON_VALUE(JSON '{"name": "Jakob", "age": "6" }', '$.age') AS scalar_age;
-- expected scalar_age: 6
```

```googlesql
SELECT JSON_VALUE('{"fruits": ["apple", "banana"]}', '$.fruits') AS json_value;
-- expected json_value: NULL (selected value is non-scalar)
```

```googlesql
SELECT JSON_VALUE('{"a.b": {"c": "world"}}', '$."a.b".c') AS hello;
-- expected hello: world
```

## Edge cases

- Returns SQL `NULL` if the JSONPath resolves to a JSON `null`.
- Returns SQL `NULL` if the JSONPath resolves to a non-scalar value such as an object or array.
- JSON keys containing characters that are invalid in JSONPath must be wrapped in double quotes within the path expression.
- Behavior may differ between the JSON-formatted `STRING` and `JSON` input types; refer to upstream notes on differences between the two.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/json_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `JSON_VALUE`

```googlesql
JSON_VALUE(json_string_expr[, json_path])
```

```googlesql
JSON_VALUE(json_expr[, json_path])
```

**Description**

Extracts a JSON scalar value and converts it to a SQL `STRING` value.
In addition, this function:

+   Removes the outermost quotes and unescapes the values.
+   Returns a SQL `NULL` if a non-scalar value is selected.
+   Uses double quotes to escape invalid [JSONPath][JSONPath-format] characters
    in JSON keys. For example: `"a.b"`.

Arguments:

+   `json_string_expr`: A JSON-formatted string. For example:

    ```
    '{"name": "Jakob", "age": "6"}'
    ```
+   `json_expr`: JSON. For example:

    ```
    JSON '{"name": "Jane", "age": "6"}'
    ```
+   `json_path`: The [JSONPath][JSONPath-format]. This identifies the data that
    you want to obtain from the input. If this optional parameter isn't
    provided, then the JSONPath `$` symbol is applied, which means that all of
    the data is analyzed.

    If `json_path` returns a JSON `null` or a non-scalar value (in other words,
    if `json_path` refers to an object or an array), then a SQL `NULL` is
    returned.

There are differences between the JSON-formatted string and JSON input types.
For details, see [Differences between the JSON and JSON-formatted STRING types][differences-json-and-string].

**Return type**

`STRING`

**Examples**

In the following example, JSON data is extracted and returned as a scalar value.

```googlesql
SELECT JSON_VALUE(JSON '{"name": "Jakob", "age": "6" }', '$.age') AS scalar_age;

/*------------+
 | scalar_age |
 +------------+
 | 6          |
 +------------*/
```

The following example compares how results are returned for the `JSON_QUERY`
and `JSON_VALUE` functions.

```googlesql
SELECT JSON_QUERY('{"name": "Jakob", "age": "6"}', '$.name') AS json_name,
  JSON_VALUE('{"name": "Jakob", "age": "6"}', '$.name') AS scalar_name,
  JSON_QUERY('{"name": "Jakob", "age": "6"}', '$.age') AS json_age,
  JSON_VALUE('{"name": "Jakob", "age": "6"}', '$.age') AS scalar_age;

/*-----------+-------------+----------+------------+
 | json_name | scalar_name | json_age | scalar_age |
 +-----------+-------------+----------+------------+
 | "Jakob"   | Jakob       | "6"      | 6          |
 +-----------+-------------+----------+------------*/
```

```googlesql
SELECT JSON_QUERY('{"fruits": ["apple", "banana"]}', '$.fruits') AS json_query,
  JSON_VALUE('{"fruits": ["apple", "banana"]}', '$.fruits') AS json_value;

/*--------------------+------------+
 | json_query         | json_value |
 +--------------------+------------+
 | ["apple","banana"] | NULL       |
 +--------------------+------------*/
```

In cases where a JSON key uses invalid JSONPath characters, you can escape those
characters using double quotes. For example:

```googlesql
SELECT JSON_VALUE('{"a.b": {"c": "world"}}', '$."a.b".c') AS hello;

/*-------+
 | hello |
 +-------+
 | world |
 +-------*/
```

[JSONPath-format]: #JSONPath_format

[differences-json-and-string]: #differences_json_and_string

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
