---
name: CHARACTER_LENGTH
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#character_length
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/character_length.yaml
---

# CHARACTER_LENGTH

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

## `CHARACTER_LENGTH`

```googlesql
CHARACTER_LENGTH(value)
```

**Description**

Synonym for [CHAR_LENGTH][string-link-to-char-length].

**Return type**

`INT64`

**Examples**

```googlesql
SELECT
  'абвгд' AS characters,
  CHARACTER_LENGTH('абвгд') AS char_length_example

/*------------+---------------------+
 | characters | char_length_example |
 +------------+---------------------+
 | абвгд      |                   5 |
 +------------+---------------------*/
```

[string-link-to-char-length]: #char_length

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
