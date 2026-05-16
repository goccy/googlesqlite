---
name: DEBUG_TOKENLIST
dialect: spanner
category: functions/search
status: implemented
notes: |
  Renders the TOKENLIST as the underlying JSON string. Runtime entry: BindSpannerDebugTokenList in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#debug_tokenlist
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#debug_tokenlist
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/debug_tokenlist.yaml
---

# DEBUG_TOKENLIST

## Summary

Returns a developer-readable string showing the tokens stored in a `TOKENLIST`. Intended for debugging index choices, **not** for use in application logic.

## Signatures

- `DEBUG_TOKENLIST(tokens)`

## Arguments

- `tokens`: `TOKENLIST`.

## Return type

`STRING`.

## Behavior

- Format is unspecified and may change between Spanner versions.
- Returns `NULL` if `tokens` is `NULL`.

## Examples

```sql
SELECT DEBUG_TOKENLIST(TOKENIZE_FULLTEXT("hello world"));
```

## Edge cases

- Production queries should not parse the output.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#debug_tokenlist>.
