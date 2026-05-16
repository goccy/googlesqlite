---
name: JSON_QUERY_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_query_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_query_array.yaml
---

# JSON_QUERY_ARRAY

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

## `JSON_QUERY_ARRAY`

```googlesql
JSON_QUERY_ARRAY(json_string_expr[, json_path])
```

```googlesql
JSON_QUERY_ARRAY(json_expr[, json_path])
```

**Description**

Extracts a JSON array and converts it to
a SQL `ARRAY<JSON-formatted STRING>` or
`ARRAY<JSON>` value.
In addition, this function uses double quotes to escape invalid
[JSONPath][JSONPath-format] characters in JSON keys. For example: `"a.b"`.

Arguments:

+   `json_string_expr`: A JSON-formatted string. For example:

    ```
    '["a", "b", {"key": "c"}]'
    ```
+   `json_expr`: JSON. For example:

    ```
    JSON '["a", "b", {"key": "c"}]'
    ```
+   `json_path`: The [JSONPath][JSONPath-format]. This identifies the data that
    you want to obtain from the input. If this optional parameter isn't
    provided, then the JSONPath `$` symbol is applied, which means that all of
    the data is analyzed.

There are differences between the JSON-formatted string and JSON input types.
For details, see [Differences between the JSON and JSON-formatted STRING types][differences-json-and-string].

**Return type**

+ `json_string_expr`: `ARRAY<JSON-formatted STRING>`
+ `json_expr`: `ARRAY<JSON>`

**Examples**

This extracts items in JSON to an array of `JSON` values:

```googlesql
SELECT JSON_QUERY_ARRAY(
  JSON '{"fruits": ["apples", "oranges", "grapes"]}', '$.fruits'
  ) AS json_array;

/*---------------------------------+
 | json_array                      |
 +---------------------------------+
 | ["apples", "oranges", "grapes"] |
 +---------------------------------*/
```

This extracts the items in a JSON-formatted string to a string array:

```googlesql
SELECT JSON_QUERY_ARRAY('[1, 2, 3]') AS string_array;

/*--------------+
 | string_array |
 +--------------+
 | [1, 2, 3]    |
 +--------------*/
```

This extracts a string array and converts it to an integer array:

```googlesql
SELECT ARRAY(
  SELECT CAST(integer_element AS INT64)
  FROM UNNEST(
    JSON_QUERY_ARRAY('[1, 2, 3]','$')
  ) AS integer_element
) AS integer_array;

/*---------------+
 | integer_array |
 +---------------+
 | [1, 2, 3]     |
 +---------------*/
```

This extracts string values in a JSON-formatted string to an array:

```googlesql
-- Doesn't strip the double quotes
SELECT JSON_QUERY_ARRAY('["apples", "oranges", "grapes"]', '$') AS string_array;

/*---------------------------------+
 | string_array                    |
 +---------------------------------+
 | ["apples", "oranges", "grapes"] |
 +---------------------------------*/
```

```googlesql
-- Strips the double quotes
SELECT ARRAY(
  SELECT JSON_VALUE(string_element, '$')
  FROM UNNEST(JSON_QUERY_ARRAY('["apples", "oranges", "grapes"]', '$')) AS string_element
) AS string_array;

/*---------------------------+
 | string_array              |
 +---------------------------+
 | [apples, oranges, grapes] |
 +---------------------------*/
```

This extracts only the items in the `fruit` property to an array:

```googlesql
SELECT JSON_QUERY_ARRAY(
  '{"fruit": [{"apples": 5, "oranges": 10}, {"apples": 2, "oranges": 4}], "vegetables": [{"lettuce": 7, "kale": 8}]}',
  '$.fruit'
) AS string_array;

/*-------------------------------------------------------+
 | string_array                                          |
 +-------------------------------------------------------+
 | [{"apples":5,"oranges":10}, {"apples":2,"oranges":4}] |
 +-------------------------------------------------------*/
```

These are equivalent:

```googlesql
SELECT JSON_QUERY_ARRAY('{"fruits": ["apples", "oranges", "grapes"]}', '$.fruits') AS string_array;

SELECT JSON_QUERY_ARRAY('{"fruits": ["apples", "oranges", "grapes"]}', '$."fruits"') AS string_array;

-- The queries above produce the following result:
/*---------------------------------+
 | string_array                    |
 +---------------------------------+
 | ["apples", "oranges", "grapes"] |
 +---------------------------------*/
```

In cases where a JSON key uses invalid JSONPath characters, you can escape those
characters using double quotes: `" "`. For example:

```googlesql
SELECT JSON_QUERY_ARRAY('{"a.b": {"c": ["world"]}}', '$."a.b".c') AS hello;

/*-----------+
 | hello     |
 +-----------+
 | ["world"] |
 +-----------*/
```

The following examples show how invalid requests and empty arrays are handled:

```googlesql
-- An error is returned if you provide an invalid JSONPath.
SELECT JSON_QUERY_ARRAY('["foo", "bar", "baz"]', 'INVALID_JSONPath') AS result;

-- If the JSONPath doesn't refer to an array, then NULL is returned.
SELECT JSON_QUERY_ARRAY('{"a": "foo"}', '$.a') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- If a key that doesn't exist is specified, then the result is NULL.
SELECT JSON_QUERY_ARRAY('{"a": "foo"}', '$.b') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/

-- Empty arrays in JSON-formatted strings are supported.
SELECT JSON_QUERY_ARRAY('{"a": "foo", "b": []}', '$.b') AS result;

/*--------+
 | result |
 +--------+
 | []     |
 +--------*/
```

[JSONPath-format]: #JSONPath_format

[differences-json-and-string]: #differences_json_and_string

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
