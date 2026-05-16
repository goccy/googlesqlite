---
name: SEARCH_NGRAMS
dialect: spanner
category: functions/search
status: implemented
notes: |
  Verifies every n-gram of the query (at the TOKENLIST's stored size) appears in the index. Runtime entry: BindSpannerSearchNGrams in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search_ngrams
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search_ngrams
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/search_ngrams.yaml
---

# SEARCH_NGRAMS

## Summary

Returns `TRUE` if a `TOKENLIST` of n-grams contains the n-grams of `query`.

## Signatures

- `SEARCH_NGRAMS(tokens, query[, min_ngrams_match, max_ngrams_query, options])`

## Arguments

- `tokens`: `TOKENLIST` produced by `TOKENIZE_NGRAMS`.
- `query`: `STRING` to search for.
- `min_ngrams_match`: optional `INT64` minimum number of n-grams that must match.
- `max_ngrams_query`: optional `INT64` upper bound on n-grams generated from the query.
- `options`: optional `STRUCT`.

## Return type

`BOOL`.

## Behavior

- Useful for fuzzy / typo-tolerant matching where exact full-text match is too strict.
- Returns `NULL` if any required argument is `NULL`.

## Examples

```sql
SELECT title FROM Books
WHERE SEARCH_NGRAMS(title_ngrams, "applied stadistics", 5);
```

## Edge cases

- Trade off recall vs precision via `min_ngrams_match`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#search_ngrams>.
