---
name: TOKENIZE_NGRAMS
dialect: spanner
category: functions/search
status: implemented
notes: |
  Tokenises into character n-grams of the requested size (default 3). Runtime entry: BindSpannerTokenizeNGrams in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_ngrams
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_ngrams
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenize_ngrams.yaml
---

# TOKENIZE_NGRAMS

## Summary

Splits a string into character n-grams (or word n-grams, per options) and returns a `TOKENLIST` for fuzzy / typo-tolerant matching.

## Signatures

- `TOKENIZE_NGRAMS(text[, ngram_size_min, ngram_size_max, options])`

## Arguments

- `text`: `STRING`.
- `ngram_size_min`, `ngram_size_max`: optional `INT64` window sizes.
- `options`: optional `STRUCT`.

## Return type

`TOKENLIST`.

## Behavior

- Returns `NULL` if `text` is `NULL`.
- Choice of `ngram_size_min`/`max` trades index size vs recall.

## Examples

```sql
SELECT TOKENIZE_NGRAMS("statistics", 2, 4);
```

## Edge cases

- Pair with `SEARCH_NGRAMS` / `SCORE_NGRAMS`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_ngrams>.
