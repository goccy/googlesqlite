---
name: SAFE_CONVERT_BYTES_TO_STRING
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#safe_convert_bytes_to_string
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/safe_convert_bytes_to_string.yaml
---

# SAFE_CONVERT_BYTES_TO_STRING

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

## `SAFE_CONVERT_BYTES_TO_STRING`

```googlesql
SAFE_CONVERT_BYTES_TO_STRING(value)
```

**Description**

Converts a sequence of `BYTES` to a `STRING`. Any invalid UTF-8 characters are
replaced with the Unicode replacement character, `U+FFFD`.

**Return type**

`STRING`

**Examples**

The following statement returns the Unicode replacement character, &#65533;.

```googlesql
SELECT SAFE_CONVERT_BYTES_TO_STRING(b'\xc2') as safe_convert;
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
