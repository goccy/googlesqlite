---
name: BAG_OF_WORDS
dialect: bigquery
category: functions/text_analysis
status: implemented
notes: |
  Self-registered through registerBigQueryExtensionFunctions.
  Runtime returns ARRAY<STRUCT<term STRING, count INT64>>, sorted
  with NULL term first then terms ascending — deterministic for the
  spec runner. NULL elements in the input are counted as a single
  NULL-term entry.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#bag_of_words
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#bag_of_words
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/text_analysis/bag_of_words.yaml
---

# BAG_OF_WORDS

## Summary
Computes the per-term frequency of a tokenized document, returning each
distinct token alongside the number of times it occurs in the input array.

## Signatures
- `BAG_OF_WORDS(tokenized_document)`

## Behavior
- `tokenized_document` is an `ARRAY<STRING>` representing a document that has
  already been tokenized into terms (tokens) for text analysis.
- The function counts how many times each distinct term appears in the input
  array and returns one row per distinct term.
- Return type is `ARRAY<STRUCT<term STRING, count INT64>>`, where `term` is a
  unique term from the input and `count` is its occurrence frequency.
- `NULL` array elements are treated as a distinct term and counted; a `NULL`
  term may appear in the result with its own count.
- Each distinct term in the input produces exactly one struct in the output
  (terms are deduplicated by counting).

## Examples
```sql
-- Per-document term frequencies
WITH ExampleTable AS (
  SELECT 1 AS id, ['I', 'like', 'pie', 'pie', 'pie', NULL] AS f UNION ALL
  SELECT 2 AS id, ['yum', 'yum', 'pie', NULL] AS f
)
SELECT id, BAG_OF_WORDS(f) AS results
FROM ExampleTable
ORDER BY id;
-- expected id=1: [(NULL,1), ('I',1), ('like',1), ('pie',3)]
-- expected id=2: [(NULL,1), ('pie',1), ('yum',2)]
```

## Edge cases
- `NULL` elements inside the array are not skipped; they are counted as a
  distinct `NULL` term in the output struct array.
- The input is expected to already be tokenized (e.g. produced by a tokenizer
  such as `TEXT_ANALYZE`); the function itself performs no tokenization or
  normalization on raw text.

## Reference (upstream)

See the upstream BigQuery documentation:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/text-analysis-functions#bag_of_words>
