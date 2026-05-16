---
name: QUOTE
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindQuote in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#quote
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#quote
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/quote.yaml
---

# QUOTE

## Summary

Returns a string that, when used as a SQL literal, evaluates to the input string.

## Signatures

- `QUOTE(str)`

## Arguments

- `str`: `STRING`.

## Return type

`STRING` enclosed in single quotes, with internal single quotes, backslashes, NULs, control characters, and other unsafe bytes escaped.

## Behavior

- A `NULL` input returns the literal four-character string `"NULL"` (without surrounding quotes).
- Backslashes and single quotes inside `str` are doubled / escaped so that re-parsing the result yields the original byte sequence.

## Examples

```sql
SELECT QUOTE("Dont");         -- "Don\\\\t"
SELECT QUOTE("a\\\\b");           -- "a\\\\\\\\b"
SELECT QUOTE(NULL);             -- "NULL"
```

## Edge cases

- Useful when assembling dynamic SQL: `CONCAT("SELECT ", QUOTE(name))` is injection-safe for the `name` interpolation.
- Output is the MySQL escaping flavor (`\\\\n`, `\\\\r`, `\\\\Z`, `\\\\0`); regenerating with a different SQL dialect may require additional handling.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#quote>.
