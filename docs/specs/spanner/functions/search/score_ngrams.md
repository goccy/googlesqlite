---
name: SCORE_NGRAMS
dialect: spanner
category: functions/search
status: implemented
notes: |
  Match-count score over n-grams; the score equals the count of query n-grams found in the index. Runtime entry: BindSpannerScoreNGrams in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#score_ngrams
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#score_ngrams
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/score_ngrams.yaml
---

# SCORE_NGRAMS

## Summary

Returns a relevance score for an n-gram tokenized document against a query.

## Signatures

- `SCORE_NGRAMS(tokens, query[, options])`

## Return type

`FLOAT64`.

## Behavior

- Companion to `SEARCH_NGRAMS` for ranked retrieval. Higher means more matching n-grams (with positional weighting per implementation).
- Returns `NULL` if any required argument is `NULL`.

## Examples

```sql
SELECT name, SCORE_NGRAMS(name_ngrams, @q) AS s
FROM Customers
WHERE SEARCH_NGRAMS(name_ngrams, @q, 3)
ORDER BY s DESC;
```

## Edge cases

- See `SCORE` regarding non-comparability of scores across queries.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#score_ngrams>.
