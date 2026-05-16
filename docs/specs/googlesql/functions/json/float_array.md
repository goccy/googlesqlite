---
name: FLOAT_ARRAY
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#float_array
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/float_array.yaml
---

# FLOAT_ARRAY

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

## `FLOAT_ARRAY` 
<a id="float_array_for_json"></a>

```googlesql
FLOAT_ARRAY(
  json_expr
  [, wide_number_mode => { 'exact' | 'round' } ]
)
```

**Description**

Converts a JSON array of numbers to a SQL `ARRAY<FLOAT>` value.

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '[9.8]'
    ```

    If the JSON value isn't an array of numbers, an error is produced. If the
    expression is a SQL `NULL`, the function returns SQL `NULL`.
+   `wide_number_mode`: A named argument with a `STRING` value. Defines
    what happens with a number that can't be represented as a
    `FLOAT` without loss of precision. This argument accepts
    one of the two case-sensitive values:

    +   `exact`: The function fails if the result can't be represented as a
        `FLOAT` without loss of precision.
    +   `round` (default): The numeric value stored in JSON will be rounded to
        `FLOAT`. If such rounding isn't possible, the function
        fails.

**Return type**

`ARRAY<FLOAT>`

**Examples**

```googlesql
SELECT FLOAT_ARRAY(JSON '[9, 9.8]') AS velocities;

/*-------------+
 | velocities  |
 +-------------+
 | [9.0, 9.8]  |
 +-------------*/
```

```googlesql
SELECT FLOAT_ARRAY(JSON '[16777217]', wide_number_mode=>'round') as result;

/*--------------+
 | result       |
 +--------------+
 | [16777216.0] |
 +--------------*/
```

```googlesql
SELECT FLOAT_ARRAY(JSON '[16777216]') as result;

/*--------------+
 | result       |
 +--------------+
 | [16777216.0] |
 +--------------*/
```

The following examples show how invalid requests are handled:

```googlesql
-- An error is thrown if the JSON isn't an array of numbers in FLOAT domain.
SELECT FLOAT_ARRAY(JSON '["strawberry"]') AS result;
SELECT FLOAT_ARRAY(JSON '[null]') AS result;
SELECT FLOAT_ARRAY(JSON 'null') AS result;

-- An error is thrown because `wide_number_mode` is case-sensitive and not "exact" or "round".
SELECT FLOAT_ARRAY(JSON '[123.4]', wide_number_mode=>'EXACT') as result;
SELECT FLOAT_ARRAY(JSON '[123.4]', wide_number_mode=>'exac') as result;

-- An error is thrown because the number can't be converted to FLOAT without loss of precision
SELECT FLOAT_ARRAY(JSON '[16777217]', wide_number_mode=>'exact') as result;
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
