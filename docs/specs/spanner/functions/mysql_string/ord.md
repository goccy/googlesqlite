---
name: ORD
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindOrd in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#ord
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#ord
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/ord.yaml
---

# ORD

## Summary

Returns the numeric value of the leftmost character of a string interpreted as a UTF-8 byte sequence (or, for plain ASCII, the code point).

## Signatures

- `ORD(str)`

## Arguments

- `str`: `STRING`.

## Return type

`INT64`.

## Behavior

- For an ASCII first character, `ORD` returns the same value as `ASCII`.
- For a multi-byte UTF-8 first character, `ORD` returns the multi-byte sequence interpreted as a big-endian integer over its bytes — for example, the two-byte UTF-8 sequence `0xC3 0x84` yields `0xC384 = 50052`.
- Returns `0` for an empty string.
- Returns `NULL` if `str` is `NULL`.

## Examples

```sql
SELECT ORD("a");     -- 97
SELECT ORD("Ä");     -- 50052
SELECT ORD("");      -- 0
```

## Edge cases

- Use `TO_CODE_POINTS` for the actual Unicode scalar value; `ORD` returns the byte-pattern interpretation, which differs from the code point for non-ASCII characters.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#ord>.
