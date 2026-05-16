---
name: CONCAT_WS
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindConcatWs in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#concat_ws
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#concat_ws
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/concat_ws.yaml
---

# CONCAT_WS

## Summary

Concatenates two or more strings with a fixed separator inserted between each pair. Unlike `CONCAT`, `NULL` arguments are skipped rather than poisoning the result.

## Signatures

- `CONCAT_WS(separator, value1, value2[, ...])`

## Arguments

- `separator`: `STRING` placed between consecutive non-`NULL` arguments. Required.
- `value1, value2, ...`: one or more `STRING` values to concatenate.

## Return type

`STRING`.

## Behavior

- `NULL` values among `value1..valueN` are skipped, not converted to the literal `"NULL"`.
- Returns `NULL` only if `separator` itself is `NULL`.
- The separator is **not** appended after the last value, nor inserted before the first value.

## Examples

```sql
SELECT CONCAT_WS(",", "a", "b", "c");           -- "a,b,c"
SELECT CONCAT_WS("-", "a", NULL, "b");          -- "a-b"
SELECT CONCAT_WS(NULL, "a", "b");               -- NULL
SELECT CONCAT_WS(",", "only");                  -- "only"
```

## Edge cases

- An empty separator produces concatenation without delimiters.
- All-`NULL` value arguments yield an empty string (not `NULL`) as long as `separator` is non-`NULL`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#concat_ws>.
