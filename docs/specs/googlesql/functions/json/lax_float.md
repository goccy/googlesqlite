---
name: LAX_FLOAT
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#lax_float
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/lax_float.yaml
---

# LAX_FLOAT

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

## `LAX_FLOAT` 
<a id="lax_float"></a>

```googlesql
LAX_FLOAT(json_expr)
```

**Description**

Attempts to convert a JSON value to a SQL `FLOAT` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '9.8'
    ```

Details:

+   If `json_expr` is SQL `NULL`, the function returns SQL `NULL`.
+   See the conversion rules in the next section for additional `NULL` handling.

**Conversion rules**

<table>
  <tr>
    <th width='200px'>From JSON type</th>
    <th>To SQL <code>FLOAT</code></th>
  </tr>
  <tr>
    <td>boolean</td>
    <td>
      <code>NULL</code>
    </td>
  </tr>
  <tr>
    <td>string</td>
    <td>
      If the JSON string represents a JSON number, parses it as
      a <code>BIGNUMERIC</code> value, and then safe casts the result as a
      <code>FLOAT</code> value.
      If the JSON string can't be converted, returns <code>NULL</code>.
    </td>
  </tr>
  <tr>
    <td>number</td>
    <td>
      Casts the JSON number as a
      <code>FLOAT</code> value.
      Large JSON numbers are rounded.
    </td>
  </tr>
  <tr>
    <td>other type or null</td>
    <td><code>NULL</code></td>
  </tr>
</table>

**Return type**

`FLOAT`

**Examples**

Examples with inputs that are JSON numbers:

```googlesql
SELECT LAX_FLOAT(JSON '9.8') AS result;

/*--------+
 | result |
 +--------+
 | 9.8    |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '9') AS result;

/*--------+
 | result |
 +--------+
 | 9.0    |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '16777217') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | 16777216.0         |
 +--------------------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '1e100') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

Examples with inputs that are JSON booleans:

```googlesql
SELECT LAX_FLOAT(JSON 'true') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON 'false') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

Examples with inputs that are JSON strings:

```googlesql
SELECT LAX_FLOAT(JSON '"10"') AS result;

/*--------+
 | result |
 +--------+
 | 10.0   |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"1.1"') AS result;

/*--------+
 | result |
 +--------+
 | 1.1    |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"1.1e2"') AS result;

/*--------+
 | result |
 +--------+
 | 110.0  |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"16777217"') AS result;

/*--------------------+
 | result             |
 +--------------------+
 | 16777216.0         |
 +--------------------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"+1.5"') AS result;

/*--------+
 | result |
 +--------+
 | 1.5    |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"NaN"') AS result;

/*--------+
 | result |
 +--------+
 | NaN    |
 +--------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"Inf"') AS result;

/*----------+
 | result   |
 +----------+
 | Infinity |
 +----------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"-InfiNiTY"') AS result;

/*-----------+
 | result    |
 +-----------+
 | -Infinity |
 +-----------*/
```

```googlesql
SELECT LAX_FLOAT(JSON '"foo"') AS result;

/*--------+
 | result |
 +--------+
 | NULL   |
 +--------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
