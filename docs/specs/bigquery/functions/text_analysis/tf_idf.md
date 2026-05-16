---
name: TF_IDF
dialect: bigquery
category: functions/text_analysis
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#tf_idf
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#tf_idf
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/text_analysis/tf_idf.yaml
---

# TF_IDF

## Summary

Window function that computes the TF-IDF relevance score of every
term in a tokenized document, ranked across the documents in the
`OVER()` window.

## Signatures

- `TF_IDF(tokenized_document) OVER()`
- `TF_IDF(tokenized_document, max_distinct_tokens) OVER()`
- `TF_IDF(tokenized_document, max_distinct_tokens, frequency_threshold) OVER()`

## Behavior

- `tokenized_document` is `ARRAY<STRING>`, typically the output of
  `TEXT_ANALYZE` or `BAG_OF_WORDS`.
- `max_distinct_tokens` (`INT64`, default 32000, max 1,048,576)
  caps the dictionary size — additional terms beyond this count
  are folded into the unknown bucket.
- `frequency_threshold` (`INT64`, default 5) is the minimum
  in-document occurrence required for a term to enter the
  dictionary.
- Returns per-row TF-IDF scores for terms in the row's tokenized
  document, evaluated across the analytic window.

## Examples

```sql
SELECT TF_IDF(tokens) OVER () AS scores
FROM (SELECT TEXT_ANALYZE(body) AS tokens FROM articles);
```

## Edge cases

- Both threshold arguments must be non-negative.
- The unknown-term bucket carries terms beyond
  `max_distinct_tokens`.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#tf_idf>.
