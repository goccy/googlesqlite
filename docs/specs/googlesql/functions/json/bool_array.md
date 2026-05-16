---
name: BOOL_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#bool_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/bool_array.yaml
---

# BOOL_ARRAY

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

## `BOOL_ARRAY` 
<a id="bool_array_for_json"></a>

```googlesql
BOOL_ARRAY(json_expr)
```

**Description**

Converts a JSON array of booleans to a SQL `ARRAY<BOOL>` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[true]'
    ```

    If the JSON value isn't an array of booleans, an error is produced. If the
    expression is SQL `NULL`, the function returns SQL `NULL`.

**Return type**

`ARRAY<BOOL>`

**Examples**

```googlesql
SELECT BOOL_ARRAY(JSON '[true, false]') AS vacancies;

/*---------------+
 | vacancies     |
 +---------------+
 | [true, false] |
 +---------------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if the JSON isn't an array of booleans.
SELECT BOOL_ARRAY(JSON '[123]') AS result; -- Throws an error
SELECT BOOL_ARRAY(JSON '[null]') AS result; -- Throws an error
SELECT BOOL_ARRAY(JSON 'null') AS result; -- Throws an error
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
