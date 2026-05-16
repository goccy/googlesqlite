---
name: FROM_BASE32
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#from_base32
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/from_base32.yaml
---

# FROM_BASE32

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

## `FROM_BASE32`

```googlesql
FROM_BASE32(string_expr)
```

**Description**

Converts the base32-encoded input `string_expr` into `BYTES` format. To convert
`BYTES` to a base32-encoded `STRING`, use [TO_BASE32][string-link-to-base32].

**Return type**

`BYTES`

**Example**

```googlesql
SELECT FROM_BASE32('MFRGGZDF74======') AS byte_data;

/*-----------+
 | byte_data |
 +-----------+
 | abcde\xff |
 +-----------*/
```

[string-link-to-base32]: #to_base32

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
