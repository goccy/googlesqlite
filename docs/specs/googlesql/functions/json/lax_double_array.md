---
name: LAX_DOUBLE_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_double_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_double_array.yaml
---

# LAX_DOUBLE_ARRAY

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

## `LAX_DOUBLE_ARRAY` 
<a id="lax_double_array"></a>

```googlesql
LAX_DOUBLE_ARRAY(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `ARRAY<DOUBLE>`
value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[9.8]'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>ARRAY&lt;DOUBLE&gt;</code></th>
  </tr>
  <tr>
    <td>array</td>
    <td>
      Converts every element according to
      <a href="#lax_double"><code>LAX_DOUBLE</code></a>
      conversion rules.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`ARRAY<DOUBLE>`

**Examples**

Examples with inputs that are JSON arrays of numbers:

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '[9.8, 9]') AS result;

/*-------------+
 | result      |
 +-------------+
 | [9.8, 9.0,] |
 +-------------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '[9007199254740993, -9007199254740993]') AS result;

/*-------------------------------------------+
 | result                                    |
 +-------------------------------------------+
 | [9007199254740992.0, -9007199254740992.0] |
 +-------------------------------------------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '[-1.79769e+308, 2.22507e-308, 1.79769e+308, 1e100]') AS result;

/*-----------------------------------------------------+
 | result                                              |
 +-----------------------------------------------------+
 | [-1.79769e+308, 2.22507e-308, 1.79769e+308, 1e+100] |
 +-----------------------------------------------------*/
```

Example with inputs that's JSON array of booleans:

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '[true, false]') AS result;

/*----------------+
 | result         |
 +----------------+
 | [NULL, NULL]   |
 +----------------*/
```

Examples with inputs that are JSON arrays of strings:

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '["10", "1.1", "1.1e2", "+1.5"]') AS result;

/*-------------------------+
 | result                  |
 +-------------------------+
 | [10.0, 1.1, 110.0, 1.5] |
 +-------------------------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '["9007199254740993"]') AS result;

/*----------------------+
 | result               |
 +----------------------+
 | [9007199254740992.0] |
 +----------------------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '["NaN", "Inf", "-InfiNiTY"]') AS result;

/*----------------------------+
 | result                     |
 +----------------------------+
 | [NaN, Infinity, -Infinity] |
 +----------------------------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '["foo", "null", ""]') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | [NULL, NULL, NULL] |
 +--------------------*/
```

Example with input that's JSON array of other types:

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '[null, {"foo": 1}, [1]]') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | [NULL, NULL, NULL] |
 +--------------------*/
```

Examples with inputs that aren't JSON arrays:

```googlesql
SELECT LAX_DOUBLE_ARRAY(NULL) AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON 'null') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_DOUBLE_ARRAY(JSON '9.8') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
