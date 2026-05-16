---
name: TRANSLATE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#translate
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/translate.yaml
---

# TRANSLATE

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

## `TRANSLATE`

```googlesql
TRANSLATE(expression, source_characters, target_characters)
```

**Description**

In `expression`, replaces each character in `source_characters` with the
corresponding character in `target_characters`. All inputs must be the same
type, either `STRING` or `BYTES`.

+ Each character in `expression` is translated at most once.
+ A character in `expression` that isn't present in `source_characters` is left
  unchanged in `expression`.
+ A character in `source_characters` without a corresponding character in
  `target_characters` is omitted from the result.
+ A duplicate character in `source_characters` results in an error.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT TRANSLATE('This is a cookie', 'sco', 'zku') AS translate

/*------------------+
 | translate        |
 +------------------+
 | Thiz iz a kuukie |
 +------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
