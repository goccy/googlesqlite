---
name: JSON_UNQUOTE
dialect: spanner
category: functions/mysql_json
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindJsonUnquote in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#json_unquote
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#json_unquote
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_json/json_unquote.yaml
---

# JSON_UNQUOTE

## Summary

Decodes a JSON string literal into the underlying `STRING` value. Inverse of `JSON_QUOTE`.

## Signatures

- `JSON_UNQUOTE(json_str)`

## Arguments

- `json_str`: `STRING` containing a JSON literal (typically a string).

## Return type

`STRING`.

## Behavior

- If `json_str` is not a JSON string literal, the value is returned unchanged.
- `\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`, and `\uXXXX` escapes are decoded.
- Returns `NULL` if the input is `NULL`.

## Examples

```sql
SELECT JSON_UNQUOTE("hello");           -- "hello"
SELECT JSON_UNQUOTE("line1\\nline2");   -- multi-line string
```

## Edge cases

- Malformed `\uXXXX` escapes raise an error.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/utility_functions#json_unquote>.
