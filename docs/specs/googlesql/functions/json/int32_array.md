---
name: INT32_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#int32_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/int32_array.yaml
---

# INT32_ARRAY

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

## `INT32_ARRAY` 
<a id="int32_array_for_json"></a>

```googlesql
INT32_ARRAY(json_expr)
```

**Description**

Converts a JSON number to a SQL `INT32_ARRAY` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[999]'
    ```

    If the JSON value isn't an array of numbers, or the JSON numbers aren't in
    the SQL `INT32` domain, an error is produced. If the expression is SQL
    `NULL`, the function returns SQL `NULL`.

**Return type**

`ARRAY<INT32>`

**Examples**

```googlesql
SELECT INT32_ARRAY(JSON '[2005, 2003]') AS flight_numbers;

/*----------------+
 | flight_numbers |
 +----------------+
 | [2005, 2003]   |
 +----------------*/
```

```googlesql
SELECT INT32_ARRAY(JSON '[10.0]') AS scores;

/*--------+
 | scores |
 +--------+
 | [10]   |
 +--------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if the JSON isn't an array of numbers in INT32 domain.
SELECT INT32_ARRAY(JSON '[10.1]') AS result;  -- Throws an error
SELECT INT32_ARRAY(JSON '["strawberry"]') AS result; -- Throws an error
SELECT INT32_ARRAY(JSON '[null]') AS result; -- Throws an error
SELECT INT32_ARRAY(JSON 'null') AS result; -- Throws an error
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
