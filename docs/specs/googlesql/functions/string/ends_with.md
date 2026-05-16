---
name: ENDS_WITH
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#ends_with
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/ends_with.yaml
---

# ENDS_WITH

## Summary

Takes two `STRING` or `BYTES` values and returns `TRUE` if `suffix` is a suffix of `value`.

## Signatures

- `ENDS_WITH(value, suffix)`

## Behavior

- Returns `BOOL`.
- Accepts two arguments of type `STRING` or `BYTES`.
- Returns `TRUE` when `suffix` is a suffix of `value`, otherwise `FALSE`.
- Supports specifying collation when comparing `STRING` values.

## Examples

```googlesql
SELECT ENDS_WITH('apple', 'e') AS example;
-- expected: True
```

## Edge cases

- Both arguments must be the same string-like type (`STRING` or `BYTES`); upstream does not document mixing.
- Collation, if specified, governs the suffix comparison for `STRING` inputs.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `ENDS_WITH`

```googlesql
ENDS_WITH(value, suffix)
```

**Description**

Takes two `STRING` or `BYTES` values. Returns `TRUE` if `suffix`
is a suffix of `value`.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return type**

`BOOL`

**Examples**

```googlesql
SELECT ENDS_WITH('apple', 'e') as example

/*---------+
 | example |
 +---------+
 |    True |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
