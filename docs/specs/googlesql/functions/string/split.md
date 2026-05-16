---
name: SPLIT
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#split
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/split.yaml
---

# SPLIT

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

## `SPLIT`

```googlesql
SPLIT(value[, delimiter])
```

**Description**

Splits a `STRING` or `BYTES` value, using a delimiter. The `delimiter` argument
must be a literal character or sequence of characters. You can't split with a
regular expression.

For `STRING`, the default delimiter is the comma `,`.

For `BYTES`, you must specify a delimiter.

Splitting on an empty delimiter produces an array of UTF-8 characters for
`STRING` values, and an array of `BYTES` for `BYTES` values.

Splitting an empty `STRING` returns an
`ARRAY` with a single empty
`STRING`.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return type**

`ARRAY<STRING>` or `ARRAY<BYTES>`

**Examples**

```googlesql
WITH letters AS
  (SELECT '' as letter_group
  UNION ALL
  SELECT 'a' as letter_group
  UNION ALL
  SELECT 'b c d' as letter_group)

SELECT SPLIT(letter_group, ' ') as example
FROM letters;

/*----------------------+
 | example              |
 +----------------------+
 | []                   |
 | [a]                  |
 | [b, c, d]            |
 +----------------------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
