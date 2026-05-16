---
name: REGEXP_REPLACE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#regexp_replace
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/regexp_replace.yaml
---

# REGEXP_REPLACE

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

## `REGEXP_REPLACE`

```googlesql
REGEXP_REPLACE(value, regexp, replacement)
```

**Description**

Returns a `STRING` where all substrings of `value` that
match regular expression `regexp` are replaced with `replacement`.

You can use backslashed-escaped digits (\1 to \9) within the `replacement`
argument to insert text matching the corresponding parenthesized group in the
`regexp` pattern. Use \0 to refer to the entire matching text.

To add a backslash in your regular expression, you must first escape it. For
example, `SELECT REGEXP_REPLACE('abc', 'b(.)', 'X\\1');` returns `aXc`. You can
also use [raw strings][string-link-to-lexical-literals] to remove one layer of
escaping, for example `SELECT REGEXP_REPLACE('abc', 'b(.)', r'X\1');`.

The `REGEXP_REPLACE` function only replaces non-overlapping matches. For
example, replacing `ana` within `banana` results in only one replacement, not
two.

If the `regexp` argument isn't a valid regular expression, this function
returns an error.

Note: GoogleSQL provides regular expression support using the
[re2][string-link-to-re2] library; see that documentation for its
regular expression syntax.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT REGEXP_REPLACE('# Heading', r'^# ([a-zA-Z0-9\s]+$)', '<h1>\\1</h1>') AS html

/*--------------------------+
 | html                     |
 +--------------------------+
 | <h1>Heading</h1>         |
 +--------------------------*/
```

[string-link-to-re2]: https://github.com/google/re2/wiki/Syntax

[string-link-to-lexical-literals]: https://github.com/google/googlesql/blob/master/docs/lexical.md#string_and_bytes_literals

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
