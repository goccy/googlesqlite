---
name: POSITION
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Returns the 1-based offset of needle in haystack (0 when not found). Catalog entry registered in registerSpannerExtensionFunctions; runtime delegate is mysql_position in internal/functions/spanner/mysql_helpers.go.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#position
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#position
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/position.yaml
---

# POSITION

## Summary

Returns the 1-based position of the first occurrence of `substr` within `str`, using the SQL-standard `POSITION(substr IN str)` syntax. Equivalent to `LOCATE(substr, str)`.

## Signatures

- `POSITION(substr IN str)`

## Arguments

- `substr`: `STRING` pattern to find.
- `str`: `STRING` to search.

## Return type

`INT64`. `0` indicates no match.

## Behavior

- Equivalent to `LOCATE(substr, str)` and `STRPOS(str, substr)` (note the inverted argument order in `STRPOS`).
- Returns `NULL` if either argument is `NULL`.

## Examples

```sql
SELECT POSITION("bar" IN "foobarbar");   -- 4
SELECT POSITION("xyz" IN "foobarbar");   -- 0
```

## Edge cases

- An empty `substr` returns `1` for any non-`NULL` `str`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#position>.
