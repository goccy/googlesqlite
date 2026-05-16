---
name: TO_BASE64
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#to_base64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/to_base64.yaml
---

# TO_BASE64

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

## `TO_BASE64`

```googlesql
TO_BASE64(bytes_expr)
```

**Description**

Converts a sequence of `BYTES` into a base64-encoded `STRING`. To convert a
base64-encoded `STRING` into `BYTES`, use [FROM_BASE64][string-link-to-from-base64].

There are several base64 encodings in common use that vary in exactly which
alphabet of 65 ASCII characters are used to encode the 64 digits and padding.
See [RFC 4648][RFC-4648] for details. This
function adds padding and uses the alphabet `[A-Za-z0-9+/=]`.

**Return type**

`STRING`

**Example**

```googlesql
SELECT TO_BASE64(b'\377\340') AS base64_string;

/*---------------+
 | base64_string |
 +---------------+
 | /+A=          |
 +---------------*/
```

To work with an encoding using a different base64 alphabet, you might need to
compose `TO_BASE64` with the `REPLACE` function. For instance, the
`base64url` url-safe and filename-safe encoding commonly used in web programming
uses `-_=` as the last characters rather than `+/=`. To encode a
`base64url`-encoded string, replace `+` and `/` with `-` and `_` respectively.

```googlesql
SELECT REPLACE(REPLACE(TO_BASE64(b'\377\340'), '+', '-'), '/', '_') as websafe_base64;

/*----------------+
 | websafe_base64 |
 +----------------+
 | _-A=           |
 +----------------*/
```

[string-link-to-from-base64]: #from_base64

[RFC-4648]: https://tools.ietf.org/html/rfc4648#section-4

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
