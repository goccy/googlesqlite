---
name: LAX_STRING_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_string_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_string_array.yaml
---

# LAX_STRING_ARRAY

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

## `LAX_STRING_ARRAY` 
<a id="lax_string_array"></a>

```googlesql
LAX_STRING_ARRAY(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `ARRAY<STRING>` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '["a", "b"]'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>STRING</code></th>
  </tr>
  <tr>
    <td>array</td>
    <td>
      Converts every element according to <a href="#lax_string"><code>LAX_STRING</code></a> conversion rules.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`ARRAY<STRING>`

**Examples**

Example with input that's a JSON array of strings:

```googlesql
SELECT LAX_STRING_ARRAY(JSON '["purple", "10"]') AS result;

/*--------------+
 | result       |
 +--------------+
 | [purple, 10] |
 +--------------*/
```

Example with input that's a JSON array of booleans:

```googlesql
SELECT LAX_STRING_ARRAY(JSON '[true, false]') AS result;

/*---------------+
 | result        |
 +---------------+
 | [true, false] |
 +---------------*/
```

Example with input that's a JSON array of numbers:

```googlesql
SELECT LAX_STRING_ARRAY(JSON '[10.0, 10, 1e100]') AS result;

/*------------------+
 | result           |
 +------------------+
 | [10, 10, 1e+100] |
 +------------------*/
```

Example with input that's a JSON array of other types:

```googlesql
SELECT LAX_STRING_ARRAY(JSON '[null, {"foo": 1}, [1]]') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | [NULL, NULL, NULL] |
 +--------------------*/
```

Examples with inputs that aren't JSON arrays:

```googlesql
SELECT LAX_STRING_ARRAY(NULL) AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_STRING_ARRAY(JSON 'null') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_STRING_ARRAY(JSON '9.8') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
