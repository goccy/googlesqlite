---
name: SOUNDEX
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#soundex
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/soundex.yaml
---

# SOUNDEX

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

## `SOUNDEX`

```googlesql
SOUNDEX(value)
```

**Description**

Returns a `STRING` that represents the
[Soundex][string-link-to-soundex-wikipedia] code for `value`.

SOUNDEX produces a phonetic representation of a string. It indexes words by
sound, as pronounced in English. It's typically used to help determine whether
two strings, such as the family names _Levine_ and _Lavine_, or the words _to_
and _too_, have similar English-language pronunciation.

The result of the SOUNDEX consists of a letter followed by 3 digits. Non-latin
characters are ignored. If the remaining string is empty after removing
non-Latin characters, an empty `STRING` is returned.

**Return type**

`STRING`

**Examples**

```googlesql
SELECT 'Ashcraft' AS value, SOUNDEX('Ashcraft') AS soundex

/*----------------------+---------+
 | value                | soundex |
 +----------------------+---------+
 | Ashcraft             | A261    |
 +----------------------+---------*/
```

[string-link-to-soundex-wikipedia]: https://en.wikipedia.org/wiki/Soundex

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
