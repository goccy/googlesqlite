---
name: REVERSE
dialect: googlesql
category: functions/string
status: implemented
source_url: docs/third_party/googlesql-docs/string_functions.md
upstream_url: https://github.com/google/googlesql/blob/master/docs/string_functions.md#reverse
last_synced: 2026-05-04
testdata: testdata/specs/googlesql/functions/string/reverse.yaml
---

# REVERSE

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

## `REVERSE`

```googlesql
REVERSE(value)
```

**Description**

Returns the reverse of the input `STRING` or `BYTES`.

**Return type**

`STRING` or `BYTES`

**Examples**

```googlesql
SELECT REVERSE('abc') AS results

/*---------+
 | results |
 +---------+
 | cba     |
 +---------*/
```

```googlesql
SELECT FORMAT('%T', REVERSE(b'1a3')) AS results

/*---------+
 | results |
 +---------+
 | b"3a1"  |
 +---------*/
```

## References

- Apache 2.0 derivative of `docs/third_party/googlesql-docs/string_functions.md`.
