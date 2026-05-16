---
name: TOKENIZE_NUMBER
dialect: spanner
category: functions/search
status: implemented
notes: |
  Stores the numeric value for SEARCH range queries. Runtime entry: BindSpannerTokenizeNumber in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_number
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_number
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenize_number.yaml
---

# TOKENIZE_NUMBER

## Summary

Builds a `TOKENLIST` from a numeric value for range-aware indexed lookups.

## Signatures

- `TOKENIZE_NUMBER(value[, options])`

## Arguments

- `value`: `INT64`, `FLOAT64`, or `NUMERIC`.
- `options`: optional `STRUCT` with bucketization parameters.

## Return type

`TOKENLIST`.

## Behavior

- The internal representation supports range predicates inside `SEARCH`.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
CREATE TABLE Products (
  ...,
  Price_T TOKENLIST AS (TOKENIZE_NUMBER(Price)) HIDDEN
) PRIMARY KEY (...);
```

## Edge cases

- For exact-match enums prefer `TOKEN`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_number>.
