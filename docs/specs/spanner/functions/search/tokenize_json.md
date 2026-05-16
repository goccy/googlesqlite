---
name: TOKENIZE_JSON
dialect: spanner
category: functions/search
status: implemented
notes: |
  Walks the JSON value and tokenises every string leaf + key. Runtime entry: BindSpannerTokenizeJSON in internal/functions/spanner/search.go (TOKENLIST is BYTES at the SQL surface).
source_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_json
upstream_url: https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_json
last_synced: 2026-05-05
testdata: testdata/specs/spanner/functions/search/tokenize_json.yaml
---

# TOKENIZE_JSON

## Summary

Tokenizes a `JSON` value, recursively flattening keys and values into a `TOKENLIST` that can be searched by path or by value.

## Signatures

- `TOKENIZE_JSON(value[, options])`

## Return type

`TOKENLIST`.

## Behavior

- Object keys, array indices, and primitive values become tokens. The exact path encoding is implementation-defined and exposed through index query syntax.
- Returns `NULL` if `value` is `NULL`.

## Examples

```sql
CREATE TABLE Events (
  ...,
  Payload_T TOKENLIST AS (TOKENIZE_JSON(Payload)) HIDDEN
) PRIMARY KEY (...);
```

## Edge cases

- Deeply nested JSON can produce very large tokenlists; bound nesting depth via `options`.

## Reference (upstream)

See <https://cloud.google.com/spanner/docs/reference/standard-sql/search_functions#tokenize_json>.
