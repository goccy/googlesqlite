---
name: BYTE_LENGTH
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#byte_length
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/byte_length.yaml
---

# BYTE_LENGTH

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

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `BYTE_LENGTH`

```googlesql
BYTE_LENGTH(value)
```

**Description**

Gets the number of `BYTES` in a `STRING` or `BYTES` value,
regardless of whether the value is a `STRING` or `BYTES` type.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT BYTE_LENGTH('абвгд') AS string_example;

/*----------------+
 | string_example |
 +----------------+
 | 10             |
 +----------------*/
```

```googlesql
SELECT BYTE_LENGTH(b'абвгд') AS bytes_example;

/*----------------+
 | bytes_example  |
 +----------------+
 | 10             |
 +----------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
