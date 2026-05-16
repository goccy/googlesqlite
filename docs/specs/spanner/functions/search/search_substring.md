---
name: SEARCH_SUBSTRING
dialect: spanner
category: functions/search
status: implemented
notes: |
  Direct substring match against the lower-cased raw payload. Runtime entry: BindSpannerSearchSubstring in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search_substring
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search_substring
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/search_substring.yaml
---

# SEARCH_SUBSTRING

## Summary

Returns `TRUE` if a substring tokenlist contains the substring tokens of `query`.

## Signatures

- `SEARCH_SUBSTRING(tokens, query[, options])`

## Arguments

- `tokens`: `TOKENLIST` produced by `TOKENIZE_SUBSTRING`.
- `query`: `STRING` substring to search for.
- `options`: optional `STRUCT`.

## Return type

`BOOL`.

## Behavior

- Match is case-sensitive by default; case folding is enabled per index options.
- Returns `NULL` if any required argument is `NULL`.

## Examples

```sql
SELECT name FROM Customers
WHERE SEARCH_SUBSTRING(name_substr, "smith");
```

## Edge cases

- Regex-style metacharacters in `query` are matched literally.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search_substring>.
