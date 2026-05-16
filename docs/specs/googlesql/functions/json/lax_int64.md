---
name: LAX_INT64
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_int64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_int64.yaml
---

# LAX_INT64

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

## `LAX_INT64` 
<a id="lax_int64"></a>

```googlesql
LAX_INT64(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `INT64` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '999'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>INT64</code></th>
  </tr>
  <tr>
    <td>boolean</td>
    <td>
      If the JSON boolean is <code>true</code>, returns <code>1</code>.
      If <code>false</code>, returns <code>0</code>.
    </td>
  </tr>
  <tr>
    <td>string</td>
    <td>
      If the JSON string represents a JSON number, parses it as
      a <code>BIGNUMERIC</code> value, and then safe casts the results as an
      <code>INT64</code> value.
      If the JSON string can't be converted, returns <code>NULL</code>.
    </td>
  </tr>
  <tr>
    <td>number</td>
    <td>
      Casts the JSON number as an <code>INT64</code> value.
      If the JSON number can't be converted, returns <code>NULL</code>.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`INT64`

**Examples**

Examples with inputs that are JSON numbers:

```googlesql
SELECT LAX_INT64(JSON '10') AS result;

/*--------+
 | result |
 +--------+
 | 10     |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '10.0') AS result;

/*--------+
 | result |
 +--------+
 | 10     |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '1.1') AS result;

/*--------+
 | result |
 +--------+
 | 1      |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '3.5') AS result;

/*--------+
 | result |
 +--------+
 | 4      |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '1.1e2') AS result;

/*--------+
 | result |
 +--------+
 | 110    |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '1e100') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

Examples with inputs that are JSON booleans:

```googlesql
SELECT LAX_INT64(JSON 'true') AS result;

/*--------+
 | result |
 +--------+
 | 1      |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON 'false') AS result;

/*--------+
 | result |
 +--------+
 | 0      |
 +--------*/
```

Examples with inputs that are JSON strings:

```googlesql
SELECT LAX_INT64(JSON '"10"') AS result;

/*--------+
 | result |
 +--------+
 | 10     |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '"1.1"') AS result;

/*--------+
 | result |
 +--------+
 | 1      |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '"1.1e2"') AS result;

/*--------+
 | result |
 +--------+
 | 110    |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '"+1.5"') AS result;

/*--------+
 | result |
 +--------+
 | 2      |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '"1e100"') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_INT64(JSON '"foo"') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
