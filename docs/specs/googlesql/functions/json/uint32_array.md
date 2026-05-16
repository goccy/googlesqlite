---
name: UINT32_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#uint32_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/uint32_array.yaml
---

# UINT32_ARRAY

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

## `UINT32_ARRAY` 
<a id="uint32_array_for_json"></a>

```googlesql
UINT32_ARRAY(json_expr)
```

**Description**

Converts a JSON number to a SQL `UINT32_ARRAY` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[999, 12]'
    ```

    If the JSON value isn't an array of numbers, or the JSON numbers aren't in
    the SQL `UINT32` domain, an error is produced. If the expression is SQL
    `NULL`, the function returns SQL `NULL`.

**Return type**

`ARRAY<UINT32>`

**Examples**

```googlesql
SELECT UINT32_ARRAY(JSON '[2005, 2003]') AS flight_numbers;

/*----------------+
 | flight_numbers |
 +----------------+
 | [2005, 2003]   |
 +----------------*/
```

```googlesql
SELECT UINT32_ARRAY(JSON '[10.0]') AS scores;

/*--------+
 | scores |
 +--------+
 | [10]   |
 +--------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if the JSON isn't an array of numbers in UINT32 domain.
SELECT UINT32_ARRAY(JSON '[10.1]') AS result;  -- Throws an error
SELECT UINT32_ARRAY(JSON '[-1]') AS result;  -- Throws an error
SELECT UINT32_ARRAY(JSON '["strawberry"]') AS result; -- Throws an error
SELECT UINT32_ARRAY(JSON '[null]') AS result; -- Throws an error
SELECT UINT32_ARRAY(JSON 'null') AS result; -- Throws an error
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
