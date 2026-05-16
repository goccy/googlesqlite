---
name: JSON_QUOTE
dialect: spanner
category: functions/mysql_json
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindJsonQuote in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#json_quote
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#json_quote
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_json/json_quote.yaml
---

# JSON_QUOTE

## Summary

Returns a JSON string literal whose decoded form is the input string. Inverse of `JSON_UNQUOTE`.

## Signatures

- `JSON_QUOTE(str)`

## Arguments

- `str`: `STRING`.

## Return type

`STRING` — a valid JSON string literal (i.e. surrounded by double quotes with internal special characters escaped).

## Behavior

- Returns `NULL` if `str` is `NULL`.

## Examples

```sql
SELECT JSON_QUOTE("hello");        -- "\"hello\""
SELECT JSON_QUOTE(line1\nline2); -- "\"line1\\nline2\""
```

## Edge cases

- Output bytes follow JSONs escaping rules (`\\\"`, `\\\\`, `\\b`, `\\f`, `\\n`, `\\r`, `\\t`, and `\\uXXXX` for other control characters).

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#json_quote>.
