---
name: STR_TO_DATE
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindStrToDate in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#str_to_date
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#str_to_date
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/str_to_date.yaml
---

# STR_TO_DATE

## Summary

Parses a string into a `DATE`/`DATETIME`/`TIMESTAMP` using a MySQL-style format string. Inverse of `DATE_FORMAT`.

## Signatures

- `STR_TO_DATE(str, format)`

## Arguments

- `str`: `STRING` to parse.
- `format`: `STRING` made of `%`-prefixed format specifiers and literal characters that must match the input.

## Return type

`DATETIME` if the format includes time fields, `DATE` otherwise.

## Behavior

- Strict matching: unmatched literal characters return `NULL`.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT STR_TO_DATE("15-03-2024", "%d-%m-%Y");   -- DATE "2024-03-15"
SELECT STR_TO_DATE("2024-03-15 14:30", "%Y-%m-%d %H:%i");
-- DATETIME "2024-03-15T14:30:00"
```

## Edge cases

- Out-of-range parsed components (e.g. `%m` = `13`) yield `NULL`.
- Use `PARSE_DATE`/`PARSE_DATETIME`/`PARSE_TIMESTAMP` for strftime-style specifiers in GoogleSQL-native code.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#str_to_date>.
