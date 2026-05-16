---
name: SEARCH
dialect: bigquery
category: functions/search
status: implemented
notes: |
  Self-registered through registerBigQueryExtensionFunctions.
  Accepts STRING / STRUCT / ARRAY<STRING> / ARRAY<STRUCT> / JSON
  inputs and recursively flattens to a token set per the chosen
  analyzer. Supported analyzers: LOG_ANALYZER (default),
  NO_OP_ANALYZER, PATTERN_ANALYZER (with regex via
  analyzer_options). json_scope honours JSON_VALUES (default),
  JSON_KEYS, JSON_KEYS_AND_VALUES. Search-index-aware
  short-circuits are intentionally not modelled — every call is
  a full scan against the tokens.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/search_functions#search
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/search_functions#search
last_synced: 2026-05-05
testdata: testdata/specs/bigquery/functions/search/search.yaml
---

# SEARCH

## Summary

Returns `TRUE` when every term in `search_query` appears in
`data_to_search` after text analysis; otherwise `FALSE`.

## Signatures

- `SEARCH(data_to_search, search_query [, json_scope => 'JSON_VALUES' | 'JSON_KEYS' | 'JSON_KEYS_AND_VALUES'] [, analyzer => 'LOG_ANALYZER' | 'NO_OP_ANALYZER' | 'PATTERN_ANALYZER'] [, analyzer_options => '...'])`

## Behavior

- `data_to_search` accepts a literal, column, list of columns, or
  table reference (treated as a `STRUCT` of columns). Returns
  `FALSE` for types other than `STRING`, `JSON`, `ARRAY<STRING>`,
  `STRUCT`, `ARRAY<STRUCT>`.
- `search_query` is a `STRING` literal or constant expression
  containing the search tokens. `NULL` raises an error.
- Optional `json_scope` controls whether JSON keys, values, or
  both are searched (default: `JSON_VALUES`).
- Optional `analyzer` selects the tokenizer
  (`LOG_ANALYZER` default, `NO_OP_ANALYZER`, `PATTERN_ANALYZER`).
- Returns `BOOL`.

## Examples

```sql
SELECT SEARCH(t, 'foo bar') FROM mydataset.mytable AS t;
-- TRUE iff every term appears in any column of the table
```

## Edge cases

- `NULL` `search_query` is an error.
- Empty token output from `LOG_ANALYZER` or `PATTERN_ANALYZER`
  raises an error.
- Search-index-aware: presence of a search index can change which
  branches are evaluated.

## Reference (upstream)

See <https://cloud.google.com/bigquery/docs/reference/standard-sql/search_functions#search>.
