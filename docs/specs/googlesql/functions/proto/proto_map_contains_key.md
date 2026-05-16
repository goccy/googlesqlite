---
name: PROTO_MAP_CONTAINS_KEY
dialect: googlesql
category: functions/proto
status: implemented
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#proto_map_contains_key
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/proto_map_contains_key.yaml
---

# PROTO_MAP_CONTAINS_KEY

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

Verbatim copy from `docs/third_party/googlesql-docs/protocol_buffer_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `PROTO_MAP_CONTAINS_KEY`

```googlesql
PROTO_MAP_CONTAINS_KEY(proto_map_field_expression, key)
```

**Description**

Returns whether a [protocol buffer map field][proto-map] contains a given key.

Input values:

+ `proto_map_field_expression`: A protocol buffer map field.
+ `key`: A key in the protocol buffer map field.

`NULL` handling:

+ If `map_field` is `NULL`, returns `NULL`.
+ If `key` is `NULL`, returns `FALSE`.

**Return type**

`BOOL`

**Examples**

To illustrate the use of this function, consider the protocol buffer message
`Item`:

```proto
message Item {
  optional map<string, int64> purchased = 1;
};
```

In the following example, the function returns `TRUE` when the key is
present, `FALSE` otherwise.

```googlesql
SELECT
  PROTO_MAP_CONTAINS_KEY(m.purchased, 'A') AS contains_a,
  PROTO_MAP_CONTAINS_KEY(m.purchased, 'B') AS contains_b
FROM
  (SELECT AS VALUE CAST("purchased { key: 'A' value: 2 }" AS Item)) AS m;

/*------------+------------+
 | contains_a | contains_b |
 +------------+------------+
 | TRUE       | FALSE      |
 +------------+------------*/
```

[proto-map]: https://developers.google.com/protocol-buffers/docs/proto3#maps

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
