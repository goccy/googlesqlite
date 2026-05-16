---
name: CHAR_LENGTH
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#char_length
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/char_length.yaml
---

# CHAR_LENGTH

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

## `CHAR_LENGTH`

```googlesql
CHAR_LENGTH(value)
```

**Description**

Gets the number of characters in a `STRING` value.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT CHAR_LENGTH('абвгд') AS char_length;

/*-------------+
 | char_length |
 +-------------+
 | 5           |
 +------------ */
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
