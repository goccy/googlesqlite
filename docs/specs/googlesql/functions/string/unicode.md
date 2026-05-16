---
name: UNICODE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#unicode
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/unicode.yaml
---

# UNICODE

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

## `UNICODE`

```googlesql
UNICODE(value)
```

**Description**

Returns the Unicode [code point][string-code-point] for the first character in
`value`. Returns `0` if `value` is empty, or if the resulting Unicode code
point is `0`.

**Return type**

`INT64`

**Examples**

```googlesql
SELECT UNICODE('âbcd') as A, UNICODE('â') as B, UNICODE('') as C, UNICODE(NULL) as D;

/*-------+-------+-------+-------+
 | A     | B     | C     | D     |
 +-------+-------+-------+-------+
 | 226   | 226   | 0     | NULL  |
 +-------+-------+-------+-------*/
```

[string-code-point]: https://en.wikipedia.org/wiki/Code_point

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
