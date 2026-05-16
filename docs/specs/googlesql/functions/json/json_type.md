---
name: JSON_TYPE
dialect: googlesql
category: functions/json
status: implemented
source_url: docs/third_party/googlesql-docs/json_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/json_functions.md#json_type
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/json/json_type.yaml
---

# JSON_TYPE

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

## `JSON_TYPE` 
<a id="json_type"></a>

```googlesql
JSON_TYPE(json_expr)
```

**Description**

Gets the JSON type of the outermost JSON value and converts the name of
this type to a SQL `STRING` value. The names of these JSON types can be
returned: `object`, `array`, `string`, `number`, `boolean`, `null`

Arguments:

+   `json_expr`: JSON. For example:

    ```
    JSON '{"name": "sky", "color": "blue"}'
    ```

    If this expression is SQL `NULL`, the function returns SQL `NULL`. If the
    extracted JSON value isn't a valid JSON type, an error is produced.

**Return type**

`STRING`

**Examples**

```googlesql
SELECT json_val, JSON_TYPE(json_val) AS type
FROM
  UNNEST(
    [
      JSON '"apple"',
      JSON '10',
      JSON '3.14',
      JSON 'null',
      JSON '{"city": "New York", "State": "NY"}',
      JSON '["apple", "banana"]',
      JSON 'false'
    ]
  ) AS json_val;

/*----------------------------------+---------+
 | json_val                         | type    |
 +----------------------------------+---------+
 | "apple"                          | string  |
 | 10                               | number  |
 | 3.14                             | number  |
 | null                             | null    |
 | {"State":"NY","city":"New York"} | object  |
 | ["apple","banana"]               | array   |
 | false                            | boolean |
 +----------------------------------+---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/json_functions.md`.
