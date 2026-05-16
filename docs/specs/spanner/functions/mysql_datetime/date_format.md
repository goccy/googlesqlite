---
name: DATE_FORMAT
dialect: spanner
category: functions/mysql_datetime
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindDateFormat in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#date_format
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#date_format
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_datetime/date_format.yaml
---

# DATE_FORMAT

## Summary

Formats a `DATE`, `DATETIME`, or `TIMESTAMP` value into a string using a MySQL-style format string.

## Signatures

- `DATE_FORMAT(value, format)`

## Arguments

- `value`: `DATE`, `DATETIME`, or `TIMESTAMP`.
- `format`: `STRING` made of literal characters and `%`-prefixed format specifiers (e.g. `%Y`, `%m`, `%d`, `%H`, `%i`, `%s`).

## Return type

`STRING`.

## Behavior

- Unknown format specifiers preserve the literal `%X` text.
- Returns `NULL` if either argument is `NULL`.

## Examples

```sql
SELECT DATE_FORMAT(DATE "2024-03-15", "%Y/%m/%d");           -- "2024/03/15"
SELECT DATE_FORMAT(TIMESTAMP "2024-03-15 12:30:45+00", "%H:%i:%s"); -- "12:30:45"
```

## Edge cases

- For GoogleSQL-native code, prefer `FORMAT_DATE`/`FORMAT_DATETIME`/`FORMAT_TIMESTAMP` (which use `strftime`-style specifiers).
- Locale-sensitive specifiers (`%W`, `%M`, `%a`, `%b`) follow English names regardless of session locale.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/datetime_functions#date_format>.
