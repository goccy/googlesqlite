---
name: TOKENIZE_SUBSTRING
dialect: spanner
category: functions/search
status: implemented
notes: |
  Stores the lower-cased raw text so SEARCH_SUBSTRING can use direct substring lookup. Runtime entry: BindSpannerTokenizeSubstring in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_substring
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_substring
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenize_substring.yaml
---

# TOKENIZE_SUBSTRING

## Summary

Tokenizes a string into a `TOKENLIST` suitable for substring search via `SEARCH_SUBSTRING`.

## Signatures

- `TOKENIZE_SUBSTRING(text[, options])`

## Return type

`TOKENLIST`.

## Behavior

- Generates an internal token representation that allows substring matching at index time without scanning every row.
- Returns `NULL` if `text` is `NULL`.

## Examples

```sql
CREATE TABLE Customers (
  ...,
  Name_Substr TOKENLIST AS (TOKENIZE_SUBSTRING(Name)) HIDDEN
) PRIMARY KEY (...);
```

## Edge cases

- Index size grows with text length more aggressively than fulltext indexing.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_substring>.
