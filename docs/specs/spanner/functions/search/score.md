---
name: SCORE
dialect: spanner
category: functions/search
status: implemented
notes: |
  Match-count score: sum of per-query-token frequencies in the TOKENLIST. Runtime entry: BindSpannerScore in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#score
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#score
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/score.yaml
---

# SCORE

## Summary

Returns a relevance score for the document, computed against the same query as a paired `SEARCH` predicate.

## Signatures

- `SCORE(tokens, query[, options])`

## Arguments

- `tokens`: `TOKENLIST` produced by `TOKENIZE_FULLTEXT`.
- `query`: `STRING` query expression.
- `options`: optional `STRUCT`.

## Return type

`FLOAT64`. Higher means more relevant. Magnitude is corpus-dependent.

## Behavior

- Score is implementation-defined (BM25-like). Use it for `ORDER BY` rather than as an absolute relevance measure.
- Returns `NULL` if any required argument is `NULL`.

## Examples

```sql
SELECT id, SCORE(tokens, @query) AS s
FROM Documents
WHERE SEARCH(tokens, @query)
ORDER BY s DESC;
```

## Edge cases

- Comparing scores between different queries or different corpora is not meaningful.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#score>.
