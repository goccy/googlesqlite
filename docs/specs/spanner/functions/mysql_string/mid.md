---
name: MID
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindMid in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#mid
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#mid
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/mid.yaml
---

# MID

## Summary

Alias of `SUBSTRING(str, pos, len)`. Returns the substring of `str` starting at 1-based `pos` and continuing for `len` characters.

## Signatures

- `MID(str, pos, len)`

## Arguments

- `str`: `STRING` to slice.
- `pos`: `INT64` 1-based start position. Negative values count from the end.
- `len`: `INT64` length of the result in characters.

## Return type

`STRING`.

## Behavior

- Behavior matches `SUBSTRING`/`SUBSTR` with three arguments.
- `pos = 0` is treated as `pos = 1` under MySQL semantics.
- `len <= 0` yields the empty string.
- Counts are by Unicode characters.

## Examples

```sql
SELECT MID("hello", 2, 3);    -- "ell"
SELECT MID("hello", -3, 2);   -- "ll"
SELECT MID("hello", 1, 0);    -- ""
```

## Edge cases

- Negative `pos` whose absolute value exceeds `CHAR_LENGTH(str)` returns the empty string.
- Provided primarily for MySQL source-code portability — prefer `SUBSTRING` / `SUBSTR` in new GoogleSQL.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#mid>.
