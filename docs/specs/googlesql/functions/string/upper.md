---
name: UPPER
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#upper
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/upper.yaml
---

# UPPER

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

## `UPPER`

```googlesql
UPPER(value)
```

**Description**

For `STRING` arguments, returns the original string with all alphabetic
characters in uppercase. Mapping between uppercase and lowercase is done
according to the
[Unicode Character Database][string-link-to-unicode-character-definitions]
without taking into account language-specific mappings.

For `BYTES` arguments, the argument is treated as ASCII text, with all bytes
greater than 127 left intact.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT UPPER('foo') AS example

/*---------+
 | example |
 +---------+
 | FOO     |
 +---------*/
```

[string-link-to-unicode-character-definitions]: http://unicode.org/ucd/

[string-link-to-strpos]: #strpos

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
