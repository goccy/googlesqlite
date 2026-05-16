---
name: UCASE
dialect: spanner
category: functions/string
status: implemented
notes: |
  Spanner-specific behaviour for this function set requires the Spanner dialect catalog (see mysql_*). Dialect plumbing is pending.
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#ucase
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#ucase
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/string/ucase.yaml
---

# UCASE

## Summary

Returns `str` with all ASCII letters upper-cased. Equivalent to `UPPER`.

## Signatures

- `UCASE(str)`

## Arguments

- `str`: `STRING`.

## Return type

`STRING`.

## Behavior

- Same as `UPPER`. Provided for MySQL portability.
- Returns `NULL` if `str` is `NULL`.

## Examples

```sql
SELECT UCASE("hello world");   -- "HELLO WORLD"
```

## Edge cases

- Locale-sensitive folding follows Unicode default; same as `UPPER`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/string_functions#ucase>.
