---
name: JSON_EXTRACT_STRING_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_extract_string_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_extract_string_array.yaml
---

# JSON_EXTRACT_STRING_ARRAY

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

## `JSON_EXTRACT_STRING_ARRAY`

Note: This function is deprecated. Consider using
[JSON_VALUE_ARRAY][json-value-array].

```googlesql
JSON_EXTRACT_STRING_ARRAY(json_string_expr[, json_path])
```

```googlesql
JSON_EXTRACT_STRING_ARRAY(json_expr[, json_path])
```

**Description**

Extracts a JSON array of scalar values and converts it to a SQL `ARRAY<STRING>`
value. In addition, this function:

+   Removes the outermost quotes and unescapes the values.
+   Returns a SQL `NULL` if the selected value isn't an array or
    not an array containing only scalar values.
+   Uses single quotes and brackets to escape invalid [JSONPath][JSONPath-format]
    characters in JSON keys. For example: `['a.b']`.

Arguments:

+   `json_string_expr`: A JSON-formatted string. For example:

    ```
    '["apples", "oranges", "grapes"]'
    ```
+   `json_expr`: JSON. For example:

    ```
    JSON '["apples", "oranges", "grapes"]'
    ```
+   `json_path`: The [JSONPath][JSONPath-format]. This identifies the data that
    you want to obtain from the input. If this optional parameter isn't
    provided, then the JSONPath `$` symbol is applied, which means that all of
    the data is analyzed.

There are differences between the JSON-formatted string and JSON input types.
For details, see [Differences between the JSON and JSON-formatted STRING types][differences-json-and-string].

Caveats:

+ A JSON `null` in the input array produces a SQL `NULL` as the output for that
  JSON `null`.
+ If a JSONPath matches an array that contains scalar objects and a JSON `null`,
  then the output is an array of the scalar objects and a SQL `NULL`.

**Return type**

`ARRAY<STRING>`

**Examples**

This extracts items in JSON to a string array:

```googlesql
SELECT JSON_EXTRACT_STRING_ARRAY(
  JSON '{"fruits": ["apples", "oranges", "grapes"]}', '$.fruits'
  ) AS string_array;

/*---------------------------+
 | string_array              |
 +---------------------------+
 | [apples, oranges, grapes] |
 +---------------------------*/
```

The following example compares how results are returned for the
`JSON_EXTRACT_ARRAY` and `JSON_EXTRACT_STRING_ARRAY` functions.

```googlesql
SELECT JSON_EXTRACT_ARRAY('["apples", "oranges"]') AS json_array,
JSON_EXTRACT_STRING_ARRAY('["apples", "oranges"]') AS string_array;

/*-----------------------+-------------------+
 | json_array            | string_array      |
 +-----------------------+-------------------+
 | ["apples", "oranges"] | [apples, oranges] |
 +-----------------------+-------------------*/
```

This extracts the items in a JSON-formatted string to a string array:

```googlesql
-- Strips the double quotes
SELECT JSON_EXTRACT_STRING_ARRAY('["foo", "bar", "baz"]', '$') AS string_array;

/*-----------------+
 | string_array    |
 +-----------------+
 | [foo, bar, baz] |
 +-----------------*/
```

This extracts a string array and converts it to an integer array:

```googlesql
SELECT ARRAY(
  SELECT CAST(integer_element AS INT64)
  FROM UNNEST(
    JSON_EXTRACT_STRING_ARRAY('[1, 2, 3]', '$')
  ) AS integer_element
) AS integer_array;

/*---------------+
 | integer_array |
 +---------------+
 | [1, 2, 3]     |
 +---------------*/
```

These are equivalent:

```googlesql
SELECT JSON_EXTRACT_STRING_ARRAY('{"fruits": ["apples", "oranges", "grapes"]}', '$[fruits]') AS string_array;

SELECT JSON_EXTRACT_STRING_ARRAY('{"fruits": ["apples", "oranges", "grapes"]}', '$.fruits') AS string_array;

-- The queries above produce the following result:
/*---------------------------+
 | string_array              |
 +---------------------------+
 | [apples, oranges, grapes] |
 +---------------------------*/
```

In cases where a JSON key uses invalid JSONPath characters, you can escape those
characters using single quotes and brackets: `[' ']`. For example:

```googlesql
SELECT JSON_EXTRACT_STRING_ARRAY('{"a.b": {"c": ["world"]}}', "$['a.b'].c") AS hello;

/*---------+
 | hello   |
 +---------+
 | [world] |
 +---------*/
```

The following examples explore how invalid requests and empty arrays are
handled:

```googlesql
-- An error is thrown if you provide an invalid JSONPath.
SELECT JSON_EXTRACT_STRING_ARRAY('["foo", "bar", "baz"]', 'INVALID_JSONPath') AS result;

-- If the JSON formatted string is invalid, then NULL is returned.
SELECT JSON_EXTRACT_STRING_ARRAY('}}', '$') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If the JSON document is NULL, then NULL is returned.
SELECT JSON_EXTRACT_STRING_ARRAY(NULL, '$') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If a JSONPath doesn't match anything, then the output is NULL.
SELECT JSON_EXTRACT_STRING_ARRAY('{"a": ["foo", "bar", "baz"]}', '$.b') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If a JSONPath matches an object that isn't an array, then the output is NULL.
SELECT JSON_EXTRACT_STRING_ARRAY('{"a": "foo"}', '$') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If a JSONPath matches an array of non-scalar objects, then the output is NULL.
SELECT JSON_EXTRACT_STRING_ARRAY('{"a": [{"b": "foo", "c": 1}, {"b": "bar", "c":2}], "d": "baz"}', '$.a') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If a JSONPath matches an array of mixed scalar and non-scalar objects, then the output is NULL.
SELECT JSON_EXTRACT_STRING_ARRAY('{"a": [10, {"b": 20}]', '$.a') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If a JSONPath matches an empty JSON array, then the output is an empty array instead of NULL.
SELECT JSON_EXTRACT_STRING_ARRAY('{"a": "foo", "b": []}', '$.b') AS result;

/*--------+
 | result |
 +--------+
 | []     |
 +--------*/

-- In the following query, the JSON null input is returned as a
-- SQL NULL in the output.
SELECT JSON_EXTRACT_STRING_ARRAY('["world", 1, null]') AS result;

/*------------------+
 | result           |
 +------------------+
 | [world, 1, NULL] |
 +------------------*/

```

[json-value-array]: #json_value_array

[JSONPath-format]: #JSONPath_format

[differences-json-and-string]: #differences_json_and_string

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
