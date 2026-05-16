---
name: TOKENIZE_FULLTEXT
dialect: spanner
category: functions/search
status: implemented
notes: |
  Splits free text into lower-cased word tokens. TOKENLIST is serialised as JSON-encoded BYTES carrying kind/tokens/raw. Runtime entry: BindSpannerTokenizeFullText in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_fulltext
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_fulltext
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenize_fulltext.yaml
---

# TOKENIZE_FULLTEXT

## Summary

Tokenizes free text for full-text search: lower-cases, splits on whitespace/punctuation, optionally stems, and yields a `TOKENLIST`.

## Signatures

- `TOKENIZE_FULLTEXT(text[, options])`

## Arguments

- `text`: `STRING` document text.
- `options`: optional `STRUCT` controlling language, stemmer, stop-word list, etc.

## Return type

`TOKENLIST`.

## Behavior

- The exact analyzer pipeline is implementation-defined and configurable through `options`.
- Returns `NULL` if `text` is `NULL`.

## Examples

```sql
SELECT TOKENIZE_FULLTEXT("Wireless headphones with active noise cancellation");
```

## Edge cases

- Pair with a generated `TOKENLIST` column for indexing.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_fulltext>.
