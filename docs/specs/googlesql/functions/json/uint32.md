---
name: UINT32
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#uint32
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/uint32.yaml
---

# UINT32

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

## `UINT32` 
<a id="uint32_for_json"></a>

```googlesql
UINT32(json_expr)
```

**Description**

Converts a JSON number to a SQL `UINT32` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '999'
    ```

    If the JSON value isn't a number, or the JSON number isn't in the SQL
    `UINT32` domain, an error is produced. If the expression is SQL `NULL`, the
    function returns SQL `NULL`.

**Return type**

`UINT32`

**Examples**

```googlesql
SELECT UINT32(JSON '2005') AS flight_number;

/*---------------+
 | flight_number |
 +---------------+
 | 2005          |
 +---------------*/
```

```googlesql
SELECT UINT32(JSON_QUERY(JSON '{"gate": "A4", "flight_number": 2005}', "$.flight_number")) AS flight_number;

/*---------------+
 | flight_number |
 +---------------+
 | 2005          |
 +---------------*/
```

```googlesql
SELECT UINT32(JSON '10.0') AS score;

/*-------+
 | score |
 +-------+
 | 10    |
 +-------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if JSON isn't a number or can't be converted to a 64-bit integer.
SELECT UINT32(JSON '10.1') AS result;  -- Throws an error
SELECT UINT32(JSON '-1') AS result;  -- Throws an error
SELECT UINT32(JSON '"strawberry"') AS result; -- Throws an error
SELECT UINT32(JSON 'null') AS result; -- Throws an error
SELECT SAFE.UINT32(JSON '"strawberry"') AS result;  -- Returns a SQL NULL
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
