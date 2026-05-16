---
name: TOKEN
dialect: spanner
category: functions/search
status: implemented
notes: |
  Wraps a single literal value into a one-element TOKENLIST. Runtime entry: BindSpannerToken in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#token
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#token
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/token.yaml
---

# TOKEN

## Summary

Constructs a single-token `TOKENLIST` from a `STRING`, suitable for indexing exact-match attribute values such as enums and identifiers.

## Signatures

- `TOKEN(value)`

## Arguments

- `value`: `STRING` (or `BOOL`/`INT64`/etc. depending on overloads — see upstream).

## Return type

`TOKENLIST` containing exactly one token.

## Behavior

- Comparison is exact: searching for a different casing or punctuation does not match.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
CREATE TABLE Orders (
  ...,
  Status_T TOKENLIST AS (TOKEN(Status)) HIDDEN
) PRIMARY KEY (...);
```

## Edge cases

- Use `TOKENIZE_FULLTEXT` for natural-language tokenization.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#token>.
