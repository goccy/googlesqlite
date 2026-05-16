---
name: LOWER
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#lower
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/lower.yaml
---

# LOWER

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

## `LOWER`

```googlesql
LOWER(value)
```

**Description**

For `STRING` arguments, returns the original string with all alphabetic
characters in lowercase. Mapping between lowercase and uppercase is done
according to the
[Unicode Character Database][string-link-to-unicode-character-definitions]
without taking into account language-specific mappings.

For `BYTES` arguments, the argument is treated as ASCII text, with all bytes
greater than 127 left intact.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT
  LOWER('FOO BAR BAZ') AS example
FROM items;

/*-------------+
 | example     |
 +-------------+
 | foo bar baz |
 +-------------*/
```

[string-link-to-unicode-character-definitions]: http://unicode.org/ucd/

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
