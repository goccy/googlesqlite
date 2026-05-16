---
name: LTRIM
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#ltrim
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/ltrim.yaml
---

# LTRIM

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

## `LTRIM`

```googlesql
LTRIM(value1[, value2])
```

**Description**

Identical to [TRIM][string-link-to-trim], but only removes leading characters.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT CONCAT('#', LTRIM('   apple   '), '#') AS example

/*-------------+
 | example     |
 +-------------+
 | #apple   #  |
 +-------------*/
```

```googlesql
SELECT LTRIM('***apple***', '*') AS example

/*-----------+
 | example   |
 +-----------+
 | apple***  |
 +-----------*/
```

```googlesql
SELECT LTRIM('xxxapplexxx', 'xyz') AS example

/*-----------+
 | example   |
 +-----------+
 | applexxx  |
 +-----------*/
```

[string-link-to-trim]: #trim

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
