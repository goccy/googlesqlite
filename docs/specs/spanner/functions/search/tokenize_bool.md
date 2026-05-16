---
name: TOKENIZE_BOOL
dialect: spanner
category: functions/search
status: implemented
notes: |
  Encodes the boolean as a single 'true'/'false' token. Runtime entry: BindSpannerTokenizeBool in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_bool
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_bool
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenize_bool.yaml
---

# TOKENIZE_BOOL

## Summary

Builds a `TOKENLIST` from a `BOOL` value, allowing boolean attribute filtering through the search index.

## Signatures

- `TOKENIZE_BOOL(value)`

## Return type

`TOKENLIST`.

## Behavior

- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
CREATE TABLE Orders (
  ...,
  Shipped_T TOKENLIST AS (TOKENIZE_BOOL(Shipped)) HIDDEN
) PRIMARY KEY (...);
```

## Edge cases

- Companion to `SEARCH(... , "true")` / `"false"` predicates.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_bool>.
