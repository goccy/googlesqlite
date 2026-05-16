---
name: PARSE_JSON
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#parse_json
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/parse_json.yaml
---

# PARSE_JSON

## Summary

Converts a JSON-formatted `STRING` value to a `JSON` value.

## Signatures

- ```googlesql
  PARSE_JSON(
    json_string_expr
    [, wide_number_mode => { 'exact' | 'round' } ]
  )
  ```

## Behavior

- Returns a `JSON` value parsed from `json_string_expr`.
- `json_string_expr` must be a valid JSON-formatted string (object, array, or scalar such as a number or quoted string).
- The optional named argument `wide_number_mode` (a `STRING`) controls how numbers that cannot be stored in `JSON` without loss of precision are handled.
- `wide_number_mode => 'exact'` is the default: a number that cannot be stored without loss of precision causes an error.
- `wide_number_mode => 'round'`: such a number is rounded to a value that can be stored without loss of precision; if it cannot be rounded, the function throws an error.
- Numbers from the following domains can be stored in JSON without loss of precision: 64-bit signed/unsigned integers (such as `INT64`) and `DOUBLE`.
- When a number appears nested inside a JSON object or array, `wide_number_mode` is applied to that nested number.

## Examples

Return type: `JSON`.

```googlesql
SELECT PARSE_JSON('{"coordinates": [10, 20], "id": 1}') AS json_data;
-- expected: {"coordinates":[10,20],"id":1}
```

```googlesql
SELECT PARSE_JSON('{"id": 922337203685477580701}', wide_number_mode=>'round') AS json_data;
-- expected: {"id":9.223372036854776e+20}
```

```googlesql
SELECT PARSE_JSON('"red"') AS json_data;
-- expected: "red"
```

## Edge cases

- Throws an error when `json_string_expr` contains a number that cannot be stored without loss of precision and `wide_number_mode` is `'exact'` (the default).
- With `wide_number_mode => 'round'`, throws an error if the over-precision number cannot be rounded to a representable value.
- Accepts JSON-formatted strings that are not name/value pairs, including bare scalars such as `'6'` or `'"red"'`.
- `wide_number_mode` only accepts the literal string values `'exact'` or `'round'`; other values are invalid.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/json_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PARSE_JSON`

```googlesql
PARSE_JSON(
  json_string_expr
  [, wide_number_mode => { 'exact' | 'round' } ]
)
```

**Description**

Converts a JSON-formatted `STRING` value to a [`JSON` value](https://www.json.org/json-en.html).

Arguments:

+   `json_string_expr`: A JSON-formatted string. For example:

    ```
    '{"class": {"students": [{"name": "Jane"}]}}'
    ```
+   `wide_number_mode`: A named argument with a `STRING` value. Determines
    how to handle numbers that can't be stored in a `JSON` value without the
    loss of precision. If used, `wide_number_mode` must include one of the
    following values:

    +   `exact` (default): Only accept numbers that can be stored without loss
        of precision. If a number that can't be stored without loss of
        precision is encountered, the function throws an error.
    +   `round`: If a number that can't be stored without loss of precision is
        encountered, attempt to round it to a number that can be stored without
        loss of precision. If the number can't be rounded, the function throws
        an error.

    If a number appears in a JSON object or array, the `wide_number_mode`
    argument is applied to the number in the object or array.

Numbers from the following domains can be stored in JSON without loss of
precision:

+ 64-bit signed/unsigned integers, such as `INT64`
+ `DOUBLE`

**Return type**

`JSON`

**Examples**

In the following example, a JSON-formatted string is converted to `JSON`.

```googlesql
SELECT PARSE_JSON('{"coordinates": [10, 20], "id": 1}') AS json_data;

/*--------------------------------+
 | json_data                      |
 +--------------------------------+
 | {"coordinates":[10,20],"id":1} |
 +--------------------------------*/
```

The following queries fail because:

+ The number that was passed in can't be stored without loss of precision.
+ `wide_number_mode=>'exact'` is used implicitly in the first query and
  explicitly in the second query.

```googlesql
SELECT PARSE_JSON('{"id": 922337203685477580701}') AS json_data; -- fails
SELECT PARSE_JSON('{"id": 922337203685477580701}', wide_number_mode=>'exact') AS json_data; -- fails
```

The following query rounds the number to a number that can be stored in JSON.

```googlesql
SELECT PARSE_JSON('{"id": 922337203685477580701}', wide_number_mode=>'round') AS json_data;

/*------------------------------+
 | json_data                    |
 +------------------------------+
 | {"id":9.223372036854776e+20} |
 +------------------------------*/
```

You can also use valid JSON-formatted strings that don't represent name/value pairs. For example:

```googlesql
SELECT PARSE_JSON('6') AS json_data;

/*------------------------------+
 | json_data                    |
 +------------------------------+
 | 6                            |
 +------------------------------*/
```

```googlesql
SELECT PARSE_JSON('"red"') AS json_data;

/*------------------------------+
 | json_data                    |
 +------------------------------+
 | "red"                        |
 +------------------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
