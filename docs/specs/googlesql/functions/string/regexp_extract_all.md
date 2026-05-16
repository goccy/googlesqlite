---
name: REGEXP_EXTRACT_ALL
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_extract_all
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_extract_all.yaml
---

# REGEXP_EXTRACT_ALL

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

## `REGEXP_EXTRACT_ALL`

```googlesql
REGEXP_EXTRACT_ALL(value, regexp)
```

**Description**

Returns an array of all substrings of `value` that match the
[re2 regular expression][string-link-to-re2], `regexp`. Returns an empty array
if there is no match.

If the regular expression contains a capturing group (`(...)`), and there is a
match for that capturing group, that match is added to the results.

The `REGEXP_EXTRACT_ALL` function only returns non-overlapping matches. For
example, using this function to extract `ana` from `banana` returns only one
substring, not two.

Returns an error if:

+ The regular expression is invalid
+ The regular expression has more than one capturing group

**Return type**

`ARRAY<STRING>` or `ARRAY<BYTES>`

**Examples**

```googlesql
SELECT REGEXP_EXTRACT_ALL('Try `func(x)` or `func(y)`', '`(.+?)`') AS example

/*--------------------+
 | example            |
 +--------------------+
 | [func(x), func(y)] |
 +--------------------*/
```

[string-link-to-re2]: https://github.com/google/re2/wiki/Syntax

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
