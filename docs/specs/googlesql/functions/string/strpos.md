---
name: STRPOS
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#strpos
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/strpos.yaml
---

# STRPOS

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

## `STRPOS`

```googlesql
STRPOS(value, subvalue)
```

**Description**

Takes two `STRING` or `BYTES` values. Returns the 1-based position of the first
occurrence of `subvalue` inside `value`. Returns `0` if `subvalue` isn't found.

This function supports specifying [collation][collation].

[collation]: https://github.com/google/googlesql/blob/master/docs/collation-concepts.md

**Return type**

`INT64`

**Examples**

```googlesql
SELECT STRPOS('foo@example.com', '@') AS example

/*---------+
 | example |
 +---------+
 |       4 |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
