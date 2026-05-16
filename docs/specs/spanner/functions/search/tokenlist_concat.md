---
name: TOKENLIST_CONCAT
dialect: spanner
category: functions/search
status: implemented
notes: |
  Merges TOKENLISTs by concatenating their token slices. Runtime entry: BindSpannerTokenListConcat in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenlist_concat
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenlist_concat
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenlist_concat.yaml
---

# TOKENLIST_CONCAT

## Summary

Concatenates multiple `TOKENLIST` values into one. Useful for building multi-field indexes.

## Signatures

- `TOKENLIST_CONCAT(tokens1, tokens2[, ...])`

## Return type

`TOKENLIST`.

## Behavior

- Tokens preserve the field origin (when supplied through generated columns) so per-field qualifiers in `SEARCH` queries continue to work.
- Returns `NULL` if any argument is `NULL`.

## Examples

```sql
CREATE TABLE Books (
  ...,
  All_Text TOKENLIST AS (
    TOKENLIST_CONCAT(
      TOKENIZE_FULLTEXT(Title),
      TOKENIZE_FULLTEXT(Description)
    )
  ) HIDDEN
) PRIMARY KEY (...);
```

## Edge cases

- Order of arguments does not affect search semantics, but it can affect score weighting per upstream documentation.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenlist_concat>.
