---
name: BOOL
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#bool
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/bool.yaml
---

# BOOL

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

## `BOOL` 
<a id="bool_for_json"></a>

```googlesql
BOOL(json_expr)
```

**Description**

Converts a JSON boolean to a SQL `BOOL` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON 'true'
    ```

    If the JSON value isn't a boolean, an error is produced. If the expression
    is SQL `NULL`, the function returns SQL `NULL`.

**Return type**

`BOOL`

**Examples**

```googlesql
SELECT BOOL(JSON 'true') AS vacancy;

/*---------+
 | vacancy |
 +---------+
 | true    |
 +---------*/
```

```googlesql
SELECT BOOL(JSON_QUERY(JSON '{"hotel class": "5-star", "vacancy": true}', "$.vacancy")) AS vacancy;

/*---------+
 | vacancy |
 +---------+
 | true    |
 +---------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if JSON isn't of type bool.
SELECT BOOL(JSON '123') AS result; -- Throws an error
SELECT BOOL(JSON 'null') AS result; -- Throws an error
SELECT SAFE.BOOL(JSON '123') AS result; -- Returns a SQL NULL
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
