---
name: PROTO_MODIFY_MAP
dialect: googlesql
category: functions/proto
status: implemented
source_url: docs/third_party/googlesql-docs/protocol_buffer_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/protocol_buffer_functions.md#proto_modify_map
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/proto/proto_modify_map.yaml
---

# PROTO_MODIFY_MAP

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

## `PROTO_MODIFY_MAP`

```googlesql
PROTO_MODIFY_MAP(proto_map_field_expression, key_value_pair[, ...])

key_value_pair:
  key, value
```

**Description**

Modifies a [protocol buffer map field][proto-map] and returns the modified map
field.

Input values:

+ `proto_map_field_expression`: A protocol buffer map field.
+ `key_value_pair`: A key-value pair in the protocol buffer map field.

Modification behavior:

+ If the key isn't already in the map field, adds the key and its value to the
  map field.
+ If the key is already in the map field, replaces its value.
+ If the key is in the map field and the value is `NULL`, removes the key and
  its value from the map field.

`NULL` handling:

+ If `key` is `NULL`, produces an error.
+ If the same `key` appears more than once, produces an error.
+ If `map` is `NULL`, `map` is treated as empty.

**Return type**

In the input protocol buffer map field, `V` as represented in `map<K,V>`.

**Examples**

To illustrate the use of this function, consider the protocol buffer message
`Item`:

```proto
message Item {
  optional map<string, int64> purchased = 1;
};
```

In the following example, the query deletes key `A`, replaces `B`, and adds
`C` in a map field called `purchased`.

```googlesql
SELECT
  PROTO_MODIFY_MAP(m.purchased, 'A', NULL, 'B', 4, 'C', 6) AS result_map
FROM
  (SELECT AS VALUE CAST("purchased { key: 'A' value: 2 } purchased { key: 'B' value: 3}" AS Item)) AS m;

/*---------------------------------------------+
 | result_map                                  |
 +---------------------------------------------+
 | { key: 'B' value: 4 } { key: 'C' value: 6 } |
 +---------------------------------------------*/
```

[proto-map]: https://developers.google.com/protocol-buffers/docs/proto3#maps

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/protocol_buffer_functions.md`.
