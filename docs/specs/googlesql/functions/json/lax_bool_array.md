---
name: LAX_BOOL_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_bool_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_bool_array.yaml
---

# LAX_BOOL_ARRAY

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

## `LAX_BOOL_ARRAY` 
<a id="lax_bool_array"></a>

```googlesql
LAX_BOOL_ARRAY(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `ARRAY<BOOL>` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[true]'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>ARRAY&lt;BOOL&gt;</code></th>
  </tr>
  <tr>
    <td>array</td>
    <td>
      Converts every element according to <a href="#lax_bool"><code>LAX_BOOL</code></a> conversion rules.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`ARRAY<BOOL>`

**Examples**

Example with input that's a JSON array of booleans:

```googlesql
SELECT LAX_BOOL_ARRAY(JSON '[true, false]') AS result;

/*---------------+
 | result        |
 +---------------+
 | [true, false] |
 +---------------*/
```

Examples with inputs that are JSON arrays of strings:

```googlesql
SELECT LAX_BOOL_ARRAY(JSON '["true", "false", "TRue", "FaLse"]') AS result;

/*----------------------------+
 | result                     |
 +----------------------------+
 | [true, false, true, false] |
 +----------------------------*/
```

```googlesql
SELECT LAX_BOOL_ARRAY(JSON '["true ", "foo", "null", ""]') AS result;

/*-------------------------+
 | result                  |
 +-------------------------+
 | [NULL, NULL, NULL, NULL |
 +-------------------------*/
```

Examples with input that's JSON array of numbers:

```googlesql
SELECT LAX_BOOL_ARRAY(JSON '[10, 0, 0.0, -1.1]') AS result;

/*--------------------------+
 | result                   |
 +--------------------------+
 | TRUE, FALSE, FALSE, TRUE |
 +--------------------------*/
```

Example with input that's JSON array of other types:

```googlesql
SELECT LAX_BOOL_ARRAY(JSON '[null, {"foo": 1}, [1]]') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | [NULL, NULL, NULL] |
 +--------------------*/
```

Examples with inputs that aren't JSON arrays:

```googlesql
SELECT LAX_BOOL_ARRAY(NULL) AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_BOOL_ARRAY(JSON 'null') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_BOOL_ARRAY(JSON 'true') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
