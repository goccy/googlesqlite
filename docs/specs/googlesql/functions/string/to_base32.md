---
name: TO_BASE32
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#to_base32
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/to_base32.yaml
---

# TO_BASE32

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

## `TO_BASE32`

```googlesql
TO_BASE32(bytes_expr)
```

**Description**

Converts a sequence of `BYTES` into a base32-encoded `STRING`. To convert a
base32-encoded `STRING` into `BYTES`, use [FROM_BASE32][string-link-to-from-base32].

**Return type**

`STRING`

**Example**

```googlesql
SELECT TO_BASE32(b'abcde\xFF') AS base32_string;

/*------------------+
 | base32_string    |
 +------------------+
 | MFRGGZDF74====== |
 +------------------*/
```

[string-link-to-from-base32]: #from_base32

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
