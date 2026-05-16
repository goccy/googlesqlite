---
name: LAX_BOOL
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_bool
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_bool.yaml
---

# LAX_BOOL

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

## `LAX_BOOL` 
<a id="lax_bool"></a>

```googlesql
LAX_BOOL(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `BOOL` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON 'true'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>BOOL</code></th>
  </tr>
  <tr>
    <td>boolean</td>
    <td>
      If the JSON boolean is <code>true</code>, returns <code>TRUE</code>.
      Otherwise, returns <code>FALSE</code>.
    </td>
  </tr>
  <tr>
    <td>string</td>
    <td>
      If the JSON string is <code>'true'</code>, returns <code>TRUE</code>.
      If the JSON string is <code>'false'</code>, returns <code>FALSE</code>.
      If the JSON string is any other value or has whitespace in it,
      returns <code>NULL</code>.
      This conversion is case-insensitive.
    </td>
  </tr>
  <tr>
    <td>number</td>
    <td>
      If the JSON number is a representation of <code>0</code>,
      returns <code>FALSE</code>. Otherwise, returns <code>TRUE</code>.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`BOOL`

**Examples**

Example with input that's a JSON boolean:

```googlesql
SELECT LAX_BOOL(JSON 'true') AS result;

/*--------+
 | result |
 +--------+
 | true   |
 +--------*/
```

Examples with inputs that are JSON strings:

```googlesql
SELECT LAX_BOOL(JSON '"true"') AS result;

/*--------+
 | result |
 +--------+
 | TRUE   |
 +--------*/
```

```googlesql
SELECT LAX_BOOL(JSON '"true "') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_BOOL(JSON '"foo"') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

Examples with inputs that are JSON numbers:

```googlesql
SELECT LAX_BOOL(JSON '10') AS result;

/*--------+
 | result |
 +--------+
 | TRUE   |
 +--------*/
```

```googlesql
SELECT LAX_BOOL(JSON '0') AS result;

/*--------+
 | result |
 +--------+
 | FALSE  |
 +--------*/
```

```googlesql
SELECT LAX_BOOL(JSON '0.0') AS result;

/*--------+
 | result |
 +--------+
 | FALSE  |
 +--------*/
```

```googlesql
SELECT LAX_BOOL(JSON '-1.1') AS result;

/*--------+
 | result |
 +--------+
 | TRUE   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
