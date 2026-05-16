---
name: HIGHWAY_FINGERPRINT128
dialect: googlesql
category: functions/hash
status: implemented
notes: |
  HighwayHash 128-bit not in the standard library and not yet vendored. Pure-Go ports exist; deferred until requested.
source_url: docs/third_party/googlesql-docs/hash_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/hash_functions.md#highway_fingerprint128
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/hash/highway_fingerprint128.yaml
---

# HIGHWAY_FINGERPRINT128

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

Verbatim copy from `docs/third_party/googlesql-docs/hash_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `HIGHWAY_FINGERPRINT128`

```
HIGHWAY_FINGERPRINT128(input STRING[, key BYTES]) -> BYTES
```

```
HIGHWAY_FINGERPRINT128(input BYTES[, key BYTES]) -> BYTES
```

**Description**

Computes the 128-bit fingerprint of a `STRING` or `BYTES` input value using the
[HighwayHash algorithm][highwayhash]. The string version treats the input as an
array of bytes. The optional `key` argument adds unpredictability to the hash
output for security.

This function returns 16 bytes.

**Return type**

BYTES

**Arguments**

*   `input`: A `STRING` or `BYTES` value to be hashed.
*   `key`: (Optional) A `BYTES` value that makes the hash output unpredictable
    to an attacker, preventing hash-flooding denial-of-service attacks. This key
    must be exactly 32 bytes (256 bits) long.

**Examples**

```googlesql
-- Without `key`

SELECT HIGHWAY_FINGERPRINT128('Hello World')

/*------------------------------------------------------------+
 | \xa3\xc3\xb8\xb4\xfeo\x88W\xbf\x88\xae\x89\xe4\xab\xd3\x8e |
 +------------------------------------------------------------*/

```

```googlesql
-- With `key`

SELECT HIGHWAY_FINGERPRINT128(
    'Hello World',
    -- 32-byte key
    b'abcdefghijklmnopqrstuvwxyzabcdef'
)

/*------------------------------------------+
 | o\x9c\xbc:tQ{p\xa5}Y\x0c\xc3\xbe\x04\xcb |
 +------------------------------------------*/

```

[highwayhash]: /docs/third_party/highwayhash/README.md?cl=head

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/hash_functions.md`.
