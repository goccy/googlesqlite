---
name: LAX_STRING
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_string.yaml
---

# LAX_STRING

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

## `LAX_STRING` 
<a id="lax_string"></a>

```googlesql
LAX_STRING(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `STRING` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '"name"'
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
    <td>boolean</td>
    <td>
      If the JSON boolean is <code>true</code>, returns <code>'true'</code>.
      If <code>false</code>, returns <code>'false'</code>.
    </td>
  </tr>
  <tr>
    <td>string</td>
    <td>
      Returns the JSON string as a <code>STRING</code> value.
    </td>
  </tr>
  <tr>
    <td>number</td>
    <td>
      Returns the JSON number as a <code>STRING</code> value.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`STRING`

**Examples**

Examples with inputs that are JSON strings:

```googlesql
SELECT LAX_STRING(JSON '"purple"') AS result;

/*--------+
 | result |
 +--------+
 | purple |
 +--------*/
```

```googlesql
SELECT LAX_STRING(JSON '"10"') AS result;

/*--------+
 | result |
 +--------+
 | 10     |
 +--------*/
```

Examples with inputs that are JSON booleans:

```googlesql
SELECT LAX_STRING(JSON 'true') AS result;

/*--------+
 | result |
 +--------+
 | true   |
 +--------*/
```

```googlesql
SELECT LAX_STRING(JSON 'false') AS result;

/*--------+
 | result |
 +--------+
 | false  |
 +--------*/
```

Examples with inputs that are JSON numbers:

```googlesql
SELECT LAX_STRING(JSON '10.0') AS result;

/*--------+
 | result |
 +--------+
 | 10     |
 +--------*/
```

```googlesql
SELECT LAX_STRING(JSON '10') AS result;

/*--------+
 | result |
 +--------+
 | 10     |
 +--------*/
```

```googlesql
SELECT LAX_STRING(JSON '1e100') AS result;

/*--------+
 | result |
 +--------+
 | 1e+100 |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
