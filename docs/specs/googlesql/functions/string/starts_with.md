---
name: STARTS_WITH
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#starts_with
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/starts_with.yaml
---

# STARTS_WITH

## Summary

Takes two `STRING` or `BYTES` values and returns `TRUE` if `prefix` is a prefix of `value`.

## Signatures

- `STARTS_WITH(value, prefix)`

## Behavior

- Returns `BOOL`.
- Accepts two `STRING` arguments or two `BYTES` arguments.
- Returns `TRUE` when `prefix` is a prefix of `value`, otherwise `FALSE`.
- Supports specifying collation when comparing `STRING` values.

## Examples

```googlesql
SELECT STARTS_WITH('bar', 'b') AS example
-- expected: True
```

## Edge cases

- Both arguments must share a type: either both `STRING` or both `BYTES`.
- Collation, when specified, governs the prefix comparison for `STRING` inputs.

## Reference (upstream)

Verbatim copy from `docs/third_party/googlesql-docs/string_functions.md`. Auto-managed by
`specctl normalize`; do not edit by hand.

## `STARTS_WITH`

```googlesql
STARTS_WITH(value, prefix)
```

**Description**

Takes two `STRING` or `BYTES` values. Returns `TRUE` if `prefix` is a
prefix of `value`.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return type**

`BOOL`

**Examples**

```googlesql
SELECT STARTS_WITH('bar', 'b') AS example

/*---------+
 | example |
 +---------+
 |    True |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
