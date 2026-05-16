---
name: STRCMP
dialect: spanner
category: functions/mysql_string
status: implemented
notes: |
  Registered in the Spanner mysql sub-catalog (registerSpannerExtensionFunctions in internal/catalog.go); runtime entry uses BindStrcmp in internal/functions/spanner/.
source_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#strcmp
upstream_url: https://cloud.google.com/spanner/docs/reference/mysql/string_functions#strcmp
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/mysql_string/strcmp.yaml
---

# STRCMP

## Summary

Returns `-1`, `0`, or `1` to indicate whether `s1` is lexicographically less than, equal to, or greater than `s2`.

## Signatures

- `STRCMP(s1, s2)`

## Arguments

- `s1`, `s2`: `STRING` values compared by their UTF-8 byte sequences (binary collation).

## Return type

`INT64` in `{-1, 0, 1}`.

## Behavior

- `s1 < s2` returns `-1`; `s1 = s2` returns `0`; `s1 > s2` returns `1`.
- Returns `NULL` if either argument is `NULL`.

## Examples

```sql
SELECT STRCMP("abc", "abd");   -- -1
SELECT STRCMP("abc", "abc");   --  0
SELECT STRCMP("abc", "abb");   --  1
```

## Edge cases

- The comparison is binary (codepoint-by-codepoint UTF-8), not locale-aware. Use `COLLATE` clauses for collation-sensitive ordering.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/mysql/string_functions#strcmp>.
