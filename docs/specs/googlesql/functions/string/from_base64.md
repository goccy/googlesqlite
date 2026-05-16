---
name: FROM_BASE64
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#from_base64
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/from_base64.yaml
---

# FROM_BASE64

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

## `FROM_BASE64`

```googlesql
FROM_BASE64(string_expr)
```

**Description**

Converts the base64-encoded input `string_expr` into
`BYTES` format. To convert
`BYTES` to a base64-encoded `STRING`,
use [TO_BASE64][string-link-to-to-base64].

There are several base64 encodings in common use that vary in exactly which
alphabet of 65 ASCII characters are used to encode the 64 digits and padding.
See [RFC 4648][RFC-4648] for details. This
function expects the alphabet `[A-Za-z0-9+/=]`.

**Return type**

`BYTES`

**Example**

```googlesql
SELECT FROM_BASE64('/+A=') AS byte_data;

/*-----------+
 | byte_data |
 +-----------+
 | \377\340  |
 +-----------*/
```

To work with an encoding using a different base64 alphabet, you might need to
compose `FROM_BASE64` with the `REPLACE` function. For instance, the
`base64url` url-safe and filename-safe encoding commonly used in web programming
uses `-_=` as the last characters rather than `+/=`. To decode a
`base64url`-encoded string, replace `-` and `_` with `+` and `/` respectively.

```googlesql
SELECT FROM_BASE64(REPLACE(REPLACE('_-A=', '-', '+'), '_', '/')) AS binary;

/*-----------+
 | binary    |
 +-----------+
 | \377\340  |
 +-----------*/
```

[RFC-4648]: https://tools.ietf.org/html/rfc4648#section-4

[string-link-to-to-base64]: #to_base64

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
