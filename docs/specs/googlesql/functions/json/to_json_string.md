---
name: TO_JSON_STRING
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#to_json_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/to_json_string.yaml
---

# TO_JSON_STRING

## Summary

Converts a SQL value to a JSON-formatted `STRING` value, optionally pretty-printed for readability.

## Signatures

- `TO_JSON_STRING(value[, pretty_print])`

## Behavior

- Returns a JSON-formatted `STRING`.
- Accepts any SQL value whose type has a defined JSON encoding in GoogleSQL.
- `pretty_print` is an optional boolean; when `true`, the returned string is formatted for easy readability.
- When `pretty_print` is omitted (or `false`), the function emits a compact JSON representation.
- When `pretty_print` is `NULL`, the function returns `NULL` regardless of the `value` argument.

## Examples

```googlesql
SELECT TO_JSON_STRING(STRUCT(1 AS id, [10,20] AS coordinates)) AS json_data;
-- expected: {"id":1,"coordinates":[10,20]}
```

```googlesql
SELECT TO_JSON_STRING(STRUCT(1 AS id, [10,20] AS coordinates), true) AS json_data;
-- expected: pretty-printed JSON with newlines and indentation for `id` and `coordinates`
```

## Edge cases

- `pretty_print = NULL` causes the function to return `NULL`, even when `value` is non-NULL.
- A `value` whose type has no defined GoogleSQL-to-JSON encoding cannot be converted; only types listed in the JSON encodings reference are supported.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/json_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `TO_JSON_STRING`

```googlesql
TO_JSON_STRING(value[, pretty_print])
```

**Description**

Converts a SQL value to a JSON-formatted `STRING` value.

Arguments:

+   `value`: A SQL value. You can review the GoogleSQL data types that
    this function supports and their JSON encodings [here][json-encodings].
+   `pretty_print`: Optional boolean parameter. If `pretty_print` is `true`, the
    returned value is formatted for easy readability. If `pretty_print` is
    `NULL`, the function returns `NULL`, regardless of the `value` argument.

**Return type**

A JSON-formatted `STRING`

**Examples**

The following query converts a `STRUCT` value to a JSON-formatted string:

```googlesql
SELECT TO_JSON_STRING(STRUCT(1 AS id, [10,20] AS coordinates)) AS json_data

/*--------------------------------+
 | json_data                      |
 +--------------------------------+
 | {"id":1,"coordinates":[10,20]} |
 +--------------------------------*/
```

The following query converts a `STRUCT` value to a JSON-formatted string that is
easy to read:

```googlesql
SELECT TO_JSON_STRING(STRUCT(1 AS id, [10,20] AS coordinates), true) AS json_data

/*--------------------+
 | json_data          |
 +--------------------+
 | {                  |
 |   "id": 1,         |
 |   "coordinates": [ |
 |     10,            |
 |     20             |
 |   ]                |
 | }                  |
 +--------------------*/
```

[json-encodings]: #json_encodings

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
