---
name: SEARCH
dialect: spanner
category: functions/search
status: implemented
notes: |
  Returns true when every query word appears as a token in the TOKENLIST. The text_analysis BindSearch detects a BYTES first-argument and routes through the Spanner runtime; STRING first-argument keeps the BigQuery path. Runtime entry: BindSearch (dispatch) in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/search.yaml
---

# SEARCH

## Summary

Returns `TRUE` if a `TOKENLIST` value matches a query string under Spanners full-text search semantics.

## Signatures

- `SEARCH(tokens, query[, options])`

## Arguments

- `tokens`: `TOKENLIST` produced by `TOKENIZE_FULLTEXT` (or stored in a search index).
- `query`: `STRING` query expression. Supports phrase quoting, `+`/`-` boolean operators, and field qualifiers depending on the index configuration.
- `options`: optional `STRUCT` controlling parser flavor and analyzer.

## Return type

`BOOL`.

## Behavior

- The query is parsed into the same token model used by the index. Tokens absent from the document or filtered by stop-word lists do not contribute.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
SELECT product_id
FROM Products
WHERE SEARCH(tokens_full, "wireless headphones");
```

## Edge cases

- An empty `query` matches no rows.
- Use `SCORE` for ranked retrieval.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search>.
