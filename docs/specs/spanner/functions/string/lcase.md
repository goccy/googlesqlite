---
name: LCASE
dialect: spanner
category: functions/string
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#lcase
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#lcase
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/string/lcase.yaml
---

# LCASE

## Summary

Returns `str` with all ASCII letters lower-cased. Equivalent to `LOWER`.

## Signatures

- `LCASE(str)`

## Arguments

- `str`: `STRING`.

## Return type

`STRING`.

## Behavior

- Same as `LOWER`. Provided for MySQL portability.
- Returns `NULL` if `str` is `NULL`.

## Examples

```sql
SELECT LCASE("Hello World");   -- "hello world"
```

## Edge cases

- Locale-sensitive folding (e.g. Turkish `İ`) follows the Unicode default; same as `LOWER`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#lcase>.
