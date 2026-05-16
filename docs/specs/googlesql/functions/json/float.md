---
name: FLOAT
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#float
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/float.yaml
---

# FLOAT

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

## `FLOAT` 
<a id="float_for_json"></a>

```googlesql
FLOAT(
  json_expr
  [, [ wide_number_mode => ] { 'exact' | 'round' } ]
)
```

**Description**

Converts a JSON number to a SQL `FLOAT` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '9.8'
    ```

    If the JSON value isn't a number, an error is produced. If the expression
    is a SQL `NULL`, the function returns SQL `NULL`.
+   `wide_number_mode`: A named argument with a `STRING` value. Defines what
    happens with a number that can't be represented as a
    `FLOAT` without loss of precision. This argument accepts
    one of the two case-sensitive values:

    +   `exact`: The function fails if the result can't be represented as a
        `FLOAT` without loss of precision.
    +   `round` (default): The numeric value stored in JSON will be rounded to
        `FLOAT`. If such rounding isn't possible, the function
        fails.

**Return type**

`FLOAT`

**Examples**

```googlesql
SELECT FLOAT(JSON '9.8') AS velocity;

/*----------+
 | velocity |
 +----------+
 | 9.8      |
 +----------*/
```

```googlesql
SELECT FLOAT(JSON_QUERY(JSON '{"vo2_max": 39.1, "age": 18}', "$.vo2_max")) AS vo2_max;

/*---------+
 | vo2_max |
 +---------+
 | 39.1    |
 +---------*/
```

```googlesql
SELECT FLOAT(JSON '16777217', wide_number_mode=>'round') as result;

/*------------+
 | result     |
 +------------+
 | 16777216.0 |
 +------------*/
```

```googlesql
SELECT FLOAT(JSON '16777216') as result;

/*------------+
 | result     |
 +------------+
 | 16777216.0 |
 +------------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if JSON isn't of type FLOAT.
SELECT FLOAT(JSON '"strawberry"') AS result;
SELECT FLOAT(JSON 'null') AS result;

-- An error is thrown because `wide_number_mode` is case-sensitive and not "exact" or "round".
SELECT FLOAT(JSON '123.4', wide_number_mode=>'EXACT') as result;
SELECT FLOAT(JSON '123.4', wide_number_mode=>'exac') as result;

-- An error is thrown because the number can't be converted to FLOAT without loss of precision
SELECT FLOAT(JSON '16777217', wide_number_mode=>'exact') as result;

-- Returns a SQL NULL
SELECT SAFE.FLOAT(JSON '"strawberry"') AS result;
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
