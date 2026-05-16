---
name: JSON_OBJECT
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_object
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_object.yaml
---

# JSON_OBJECT

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

## `JSON_OBJECT`

+   [Signature 1](#json_object_signature1):
    `JSON_OBJECT([json_key, json_value][, ...])`
+   [Signature 2](#json_object_signature2):
    `JSON_OBJECT(json_key_array, json_value_array)`

#### Signature 1 
<a id="json_object_signature1"></a>

```googlesql
JSON_OBJECT([json_key, json_value][, ...])
```

**Description**

Creates a JSON object, using key-value pairs.

Arguments:

+   `json_key`: A `STRING` value that represents a key.
+   `json_value`: A [JSON encoding-supported][json-encodings] value.

Details:

+   If two keys are passed in with the same name, only the first key-value pair
    is preserved.
+   The order of key-value pairs isn't preserved.
+   If `json_key` is `NULL`, an error is produced.

**Return type**

`JSON`

**Examples**

You can create an empty JSON object by passing in no JSON keys and values.
For example:

```googlesql
SELECT JSON_OBJECT() AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {}        |
 +-----------*/
```

You can create a JSON object by passing in key-value pairs. For example:

```googlesql
SELECT JSON_OBJECT('foo', 10, 'bar', TRUE) AS json_data

/*-----------------------+
 | json_data             |
 +-----------------------+
 | {"bar":true,"foo":10} |
 +-----------------------*/
```

```googlesql
SELECT JSON_OBJECT('foo', 10, 'bar', ['a', 'b']) AS json_data

/*----------------------------+
 | json_data                  |
 +----------------------------+
 | {"bar":["a","b"],"foo":10} |
 +----------------------------*/
```

```googlesql
SELECT JSON_OBJECT('a', NULL, 'b', JSON 'null') AS json_data

/*---------------------+
 | json_data           |
 +---------------------+
 | {"a":null,"b":null} |
 +---------------------*/
```

```googlesql
SELECT JSON_OBJECT('a', 10, 'a', 'foo') AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {"a":10}  |
 +-----------*/
```

```googlesql
WITH Items AS (SELECT 'hello' AS key, 'world' AS value)
SELECT JSON_OBJECT(key, value) AS json_data FROM Items

/*-------------------+
 | json_data         |
 +-------------------+
 | {"hello":"world"} |
 +-------------------*/
```

An error is produced if a SQL `NULL` is passed in for a JSON key.

```googlesql
-- Error: A key can't be NULL.
SELECT JSON_OBJECT(NULL, 1) AS json_data
```

An error is produced if the number of JSON keys and JSON values don't match:

```googlesql
-- Error: No matching signature for function JSON_OBJECT for argument types:
-- STRING, INT64, STRING
SELECT JSON_OBJECT('a', 1, 'b') AS json_data
```

#### Signature 2 
<a id="json_object_signature2"></a>

```googlesql
JSON_OBJECT(json_key_array, json_value_array)
```

Creates a JSON object, using an array of keys and values.

Arguments:

+   `json_key_array`: An array of zero or more `STRING` keys.
+   `json_value_array`: An array of zero or more
    [JSON encoding-supported][json-encodings] values.

Details:

+   If two keys are passed in with the same name, only the first key-value pair
    is preserved.
+   The order of key-value pairs isn't preserved.
+   The number of keys must match the number of values, otherwise an error is
    produced.
+   If any argument is `NULL`, an error is produced.
+   If a key in `json_key_array` is `NULL`, an error is produced.

**Return type**

`JSON`

**Examples**

You can create an empty JSON object by passing in an empty array of
keys and values. For example:

```googlesql
SELECT JSON_OBJECT(CAST([] AS ARRAY<STRING>), []) AS json_data

/*-----------+
 | json_data |
 +-----------+
 | {}        |
 +-----------*/
```

You can create a JSON object by passing in an array of keys and an array of
values. For example:

```googlesql
SELECT JSON_OBJECT(['a', 'b'], [10, NULL]) AS json_data

/*-------------------+
 | json_data         |
 +-------------------+
 | {"a":10,"b":null} |
 +-------------------*/
```

```googlesql
SELECT JSON_OBJECT(['a', 'b'], [JSON '10', JSON '"foo"']) AS json_data

/*--------------------+
 | json_data          |
 +--------------------+
 | {"a":10,"b":"foo"} |
 +--------------------*/
```

```googlesql
SELECT
  JSON_OBJECT(
    ['a', 'b'],
    [STRUCT(10 AS id, 'Red' AS color), STRUCT(20 AS id, 'Blue' AS color)])
    AS json_data

/*------------------------------------------------------------+
 | json_data                                                  |
 +------------------------------------------------------------+
 | {"a":{"color":"Red","id":10},"b":{"color":"Blue","id":20}} |
 +------------------------------------------------------------*/
```

```googlesql
SELECT
  JSON_OBJECT(
    ['a', 'b'],
    [TO_JSON(10), TO_JSON(['foo', 'bar'])])
    AS json_data

/*----------------------------+
 | json_data                  |
 +----------------------------+
 | {"a":10,"b":["foo","bar"]} |
 +----------------------------*/
```

The following query groups by `id` and then creates an array of keys and
values from the rows with the same `id`:

```googlesql
WITH
  Fruits AS (
    SELECT 0 AS id, 'color' AS json_key, 'red' AS json_value UNION ALL
    SELECT 0, 'fruit', 'apple' UNION ALL
    SELECT 1, 'fruit', 'banana' UNION ALL
    SELECT 1, 'ripe', 'true'
  )
SELECT JSON_OBJECT(ARRAY_AGG(json_key), ARRAY_AGG(json_value)) AS json_data
FROM Fruits
GROUP BY id

/*----------------------------------+
 | json_data                        |
 +----------------------------------+
 | {"color":"red","fruit":"apple"}  |
 | {"fruit":"banana","ripe":"true"} |
 +----------------------------------*/
```

An error is produced if the size of the JSON keys and values arrays don't
match:

```googlesql
-- Error: The number of keys and values must match.
SELECT JSON_OBJECT(['a', 'b'], [10]) AS json_data
```

An error is produced if the array of JSON keys or JSON values is a SQL `NULL`.

```googlesql
-- Error: The keys array can't be NULL.
SELECT JSON_OBJECT(CAST(NULL AS ARRAY<STRING>), [10, 20]) AS json_data
```

```googlesql
-- Error: The values array can't be NULL.
SELECT JSON_OBJECT(['a', 'b'], CAST(NULL AS ARRAY<INT64>)) AS json_data
```

[json-encodings]: #json_encodings

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
