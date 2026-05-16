---
name: INSERT
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindInsert in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#insert
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#insert
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/insert.yaml
---

# INSERT

## Summary

Replaces a substring of `str` starting at position `pos` and `len` characters long with `newstr`, returning the result.

## Signatures

- `INSERT(str, pos, len, newstr)`

## Arguments

- `str`: `STRING` to operate on.
- `pos`: `INT64` 1-based position of the first character to replace.
- `len`: `INT64` number of characters to remove from `str` starting at `pos`. Negative or out-of-range values are clamped per MySQL semantics.
- `newstr`: `STRING` inserted in place of the removed segment.

## Return type

`STRING`.

## Behavior

- `pos < 1` or `pos > CHAR_LENGTH(str)` returns `str` unchanged.
- If `pos + len` exceeds the end of `str`, the rest of the string from `pos` onward is replaced.
- Counts are by Unicode characters, not bytes.
- Any `NULL` argument propagates and yields `NULL`.

## Examples

```sql
SELECT INSERT("Quadratic", 3, 4, "What");  -- "QuWhattic"
SELECT INSERT("Quadratic", -1, 4, "What"); -- "Quadratic"
SELECT INSERT("Quadratic", 3, 100, "X");   -- "QuX"
```

## Edge cases

- An empty `newstr` performs a deletion rather than an insertion.
- For `BYTES`, callers should use `BYTES`-specific operations; `INSERT` is defined on `STRING`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#insert>.
