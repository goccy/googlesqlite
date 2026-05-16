---
name: RTRIM
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#rtrim
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/rtrim.yaml
---

# RTRIM

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

## `RTRIM`

```googlesql
RTRIM(value1[, value2])
```

**Description**

Identical to [TRIM][string-link-to-trim], but only removes trailing characters.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT RTRIM('***apple***', '*') AS example

/*-----------+
 | example   |
 +-----------+
 | ***apple  |
 +-----------*/
```

```googlesql
SELECT RTRIM('applexxz', 'xyz') AS example

/*---------+
 | example |
 +---------+
 | apple   |
 +---------*/
```

[string-link-to-trim]: #trim

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
