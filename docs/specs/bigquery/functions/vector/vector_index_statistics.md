---
name: VECTOR_INDEX.STATISTICS
dialect: bigquery
category: functions/vector
status: implemented
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/vectorindex_functions#vector_indexstatistics
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/vectorindex_functions#vector_indexstatistics
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/vector/vector_index_statistics.yaml
---

# VECTOR_INDEX.STATISTICS

## Summary

Table-valued function that reports how much an indexed table has
drifted from the data the active vector index was trained on.

## Signatures

- `VECTOR_INDEX.STATISTICS(TABLE table_name)`

## Behavior

- `table_name` identifies a table (`dataset.table` form) that
  carries a vector index.
- Returns a single-column row with a `FLOAT64` drift score in
  `[0, 1)`. Lower is less drift; ≥ 0.3 is typically significant.
- Returns empty results when the table has no active vector index.
- Returns a `NULL` drift score when the active index is still
  training.
- Caller needs `roles/bigquery.dataEditor` or
  `roles/bigquery.dataOwner` on the table.

## Examples

```sql
SELECT * FROM VECTOR_INDEX.STATISTICS(TABLE mydataset.mytable);
```

## Edge cases

- No active index → empty output.
- Index in training → drift score is `NULL`.
- Insufficient IAM permissions raise an error before the function
  runs.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/vectorindex_functions#vector_indexstatistics>.
