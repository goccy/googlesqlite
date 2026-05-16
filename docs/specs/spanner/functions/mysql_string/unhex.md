---
name: UNHEX
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindUnhex in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#unhex
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#unhex
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/unhex.yaml
---

# UNHEX

## Summary

Decodes a string of hex digits into the corresponding `BYTES` value. Inverse of `HEX` for byte/string inputs.

## Signatures

- `UNHEX(str)`

## Arguments

- `str`: `STRING` containing case-insensitive hex digits `0-9A-Fa-f`. Any other character causes an error.

## Return type

`BYTES`.

## Behavior

- An odd-length input is treated as if a leading `"0"` were prepended (consistent with MySQL).
- Whitespace inside `str` is **not** ignored; trim it out first.
- Returns `NULL` if `str` is `NULL`.

## Examples

```sql
SELECT UNHEX("4D7953514C");        -- b"MySQL"
SELECT UNHEX("FF");                -- b"\\xff"
SELECT UNHEX("0");                 -- b"\\x00"
```

## Edge cases

- For round-tripping integers, `CAST(UNHEX(HEX(x)) AS INT64)` works only when the hex representation has exactly 16 characters.
- Use `FROM_HEX` for the GoogleSQL-native equivalent — `UNHEX` is provided for MySQL portability.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#unhex>.
