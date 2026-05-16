---
name: VECTOR_SEARCH
dialect: bigquery
category: functions/search
status: unsupported
notes: |
  Requires a vector index (HNSW / IVF) over an underlying table. googlesqlite does not maintain vector indexes; until a brute-force fallback is requested by the consumer, the function stays unsupported.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/search_functions#vector_search
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/search_functions#vector_search
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/search/vector_search.yaml
---

# VECTOR_SEARCH

## Summary

Table-valued function that finds the nearest-neighbour embeddings
in a base table relative to one or more query vectors.

## Signatures

- Batch (multi-row queries):
  `VECTOR_SEARCH({TABLE base_table | (base_table_query)}, column_to_search, {TABLE query_table | (query_table_query)} [, query_column_to_search => ...] [, top_k => N] [, distance_type => 'EUCLIDEAN'|'COSINE'|...] [, options => '...'])`
- Single-vector (Preview):
  `VECTOR_SEARCH({TABLE base_table | (base_table_query)}, column_to_search, query_value => single_query_value [, top_k => N] [, distance_type => ...] [, options => ...])`

## Behavior

- `base_table` is the table whose `column_to_search` holds the
  embedding vectors. Pre-filter via a `base_table_query` containing
  only `SELECT / FROM / WHERE` if needed.
- `column_to_search` must be of an allowed embedding type
  (e.g. `ARRAY<FLOAT64>`).
- `top_k` (default 10) caps the number of nearest neighbours
  returned per query row.
- `distance_type` selects the distance metric (default
  `EUCLIDEAN`).
- The single-vector form is optimised for one query at a time and
  performs better than the batch form on a one-row query table.

## Examples

```sql
SELECT *
FROM VECTOR_SEARCH(
  TABLE mydataset.embeddings, 'embedding',
  TABLE mydataset.queries,
  top_k => 5,
  distance_type => 'COSINE'
);
```

## Edge cases

- Filtering on non-indexed columns in `base_table_query` causes
  post-filtering instead of pruning, lowering performance.
- `WHERE` clauses on the embedding column itself are forbidden.
- Logical views are not allowed in `base_table_query`.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/search_functions#vector_search>.
