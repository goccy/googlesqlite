---
name: LAX_UINT32_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_uint32_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_uint32_array.yaml
---

# LAX_UINT32_ARRAY

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

## `LAX_UINT32_ARRAY` 
<a id="lax_uint32_array"></a>

```googlesql
LAX_UINT32_ARRAY(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `ARRAY<UINT32>` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[999, 12]'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>ARRAY&lt;UINT32&gt;</code></th>
  </tr>
  <tr>
    <td>array</td>
    <td>
      Converts every element according to
      <a href="#lax_uint32"><code>LAX_UINT32</code></a>
      conversion rules.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`ARRAY<UINT32>`

**Examples**

Examples with inputs that are JSON arrays of numbers:

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '[10, 10.0, 1.1, 3.5, 1.1e2]') AS result;

/*---------------------+
 | result              |
 +---------------------+
 | [10, 10, 1, 4, 110] |
 +---------------------*/
```

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '[1e100]') AS result;

/*--------+
 | result |
 +--------+
 | [NULL] |
 +--------*/
```

Example with inputs that's a JSON array of booleans:

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '[true, false]') AS result;

/*--------+
 | result |
 +--------+
 | [1, 0] |
 +--------*/
```

Examples with inputs that are JSON strings:

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '["10", "1.1", "1.1e2", "+1.5"]') AS result;

/*-----------------+
 | result          |
 +-----------------+
 | [10, 1, 110, 2] |
 +-----------------*/
```

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '["1e100"]') AS result;

/*--------+
 | result |
 +--------+
 | [NULL] |
 +--------*/
```

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '["foo", "null", ""]') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | [NULL, NULL, NULL] |
 +--------------------*/
```

Example with input that's a JSON array of other types:

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '[null, {"foo": 1}, [1]]') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | [NULL, NULL, NULL] |
 +--------------------*/
```

Examples with inputs that aren't JSON arrays:

```googlesql
SELECT LAX_UINT32_ARRAY(NULL) AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_UINT32_ARRAY(JSON 'null') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_UINT32_ARRAY(JSON '9.8') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
